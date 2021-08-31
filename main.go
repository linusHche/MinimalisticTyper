package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jroimartin/gocui"
)

type KeyStroke struct {
	key  gocui.Key
	char rune
}

type Word struct {
	text  string
	index int
}

type TypingRound struct {
	nextWord         chan bool
	keyStroke        chan KeyStroke
	originalPrompt   string
	words            []Word
	currentWordIndex int
	inputCorrect     bool
	hasStarted       bool
	data             []float64
}

type System struct {
	newPassage       chan string
	completedPassage chan bool
	tr               *TypingRound
	g                *gocui.Gui
	passages         []string
}

var file, _ = os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)

func main() {

	filePath := os.Args[1]

	data, _ := os.ReadFile(filePath)

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	log.SetOutput(file)

	ps := strings.Split(string(data), "\n")
	s := System{
		newPassage:       make(chan string),
		completedPassage: make(chan bool, 3),
		g:                g,
		tr:               makeTypingRound(getRandomPassage(ps)),
		passages:         ps,
	}

	g.SetManagerFunc(s.typingRoundLayout)
	go s.handleViews(g)
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func (s *System) handleViews(g *gocui.Gui) {
	for {
		select {
		case p := <-s.newPassage:
			g.Update(func(g *gocui.Gui) error {
				s.tr = makeTypingRound(p)

				g.DeleteView("paragraph")
				g.DeleteView("WPM")
				g.DeleteView("input")
				g.DeleteView("stats")
				g.DeleteView("averageWPM")
				g.DeleteView("continue")
				g.DeleteView("skip")
				return s.renderTypingView()
			})
		case c := <-s.completedPassage:
			if c {
				g.Update(func(g *gocui.Gui) error {
					return s.renderStatsView()
				})
			}
		}
	}
}

func (s *System) typingRoundLayout(g *gocui.Gui) error {
	return s.renderTypingView()
}

func (s *System) handleTypingInputs() {

outer:
	for {
		select {
		case t := <-s.tr.keyStroke:
			s.g.Update(func(g *gocui.Gui) error {
				v, _ := g.View("paragraph")
				return s.updatePromptView(v, t.key)
			})
		case <-s.completedPassage:
			break outer
		}
	}

}

func (s *System) updatePromptView(v *gocui.View, input gocui.Key) error {
	tr := s.tr
	wordOutput := make([]string, len(tr.words))

	for i, w := range tr.words {
		wordOutput[i] = w.text
	}

	canMoveToNextWord := s.checkIfCanMoveOnToNextWord()

	if input == gocui.KeySpace {
		if canMoveToNextWord {
			tr.currentWordIndex++
			tr.nextWord <- true
			if tr.currentWordIndex == len(tr.words) {
				s.completedPassage <- true
				s.completedPassage <- true
				s.completedPassage <- true
				return nil
			}
		}
	}

	wordOutput[tr.currentWordIndex] = markWord(wordOutput[tr.currentWordIndex], s.isCorrectSoFar() || canMoveToNextWord)

	v.Clear()
	fmt.Fprintln(v, strings.Join(wordOutput, " "))
	return nil
}

func (s *System) recordTypingStats() {
	tr := s.tr
	start := time.Now()
	tr.data = append(s.tr.data, 0)
outer:
	for {
		select {
		case <-time.After(time.Second / 2):
			if tr.currentWordIndex < len(tr.words) {
				s.g.Update(func(g *gocui.Gui) error {
					wv, _ := g.View("WPM")
					chars := tr.words[tr.currentWordIndex].index
					t := time.Now()
					curWPM := float64(chars) / t.Sub(start).Seconds() / 5 * 60
					tr.data = append(tr.data, curWPM)
					wv.Clear()
					fmt.Fprintln(wv, int(curWPM))
					return nil
				})
			}
		case <-s.completedPassage:
			break outer
		}
	}
}

func (s *System) typingEditor(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {
	tr := s.tr
	if !tr.hasStarted {
		tr.hasStarted = true
		go s.recordTypingStats()
	}

	ks := KeyStroke{
		key:  key,
		char: ch,
	}
	tr.keyStroke <- ks
	switch key {
	case gocui.KeyBackspace2:
		v.EditDelete(true)
	case gocui.KeyArrowLeft:
		v.MoveCursor(-1, 0, false)
	case gocui.KeyArrowRight:
		v.MoveCursor(1, 0, false)
	case gocui.KeyBackspace:
		v.Clear()
		v.SetCursor(0, 0)
	case gocui.KeySpace:
		go func() {
			nw := <-tr.nextWord
			if nw {
				s.g.Update(func(g *gocui.Gui) error {
					v, _ := g.View("input")
					v.SetCursor(0, 0)
					v.Clear()
					return nil
				})
				return
			}
		}()
	default:
		v.EditWrite(ch)
	}
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
