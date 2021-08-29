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
	gui              *gocui.Gui
	passages         []string
}

var passage string = "Born too late to explore the Earth. Born too soon to explore the universe. Born just in time to code in Golang."

// var passage string = "Born too"
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

	s := System{
		newPassage:       make(chan string),
		completedPassage: make(chan bool, 3),
		gui:              g,
		tr: &TypingRound{
			nextWord:         make(chan bool),
			keyStroke:        make(chan KeyStroke),
			originalPrompt:   passage,
			words:            processPrompt(passage),
			currentWordIndex: 0,
			inputCorrect:     false,
			hasStarted:       false,
		},
		passages: strings.Split(string(data), "\n"),
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
				s.tr.originalPrompt = p
				s.tr.words = processPrompt(p)
				s.tr.currentWordIndex = 0
				s.tr.inputCorrect = false
				s.tr.hasStarted = false
				s.tr.data = s.tr.data[:0]
				g.DeleteView("paragraph")
				g.DeleteView("averageWPM")
				g.DeleteView("continue")
				g.DeleteView("WPM")
				g.DeleteView("input")
				g.DeleteView("stats")
				return s.renderTypingView(g, p, s.completedPassage)
			})
		case <-s.completedPassage:
			g.Update(func(g *gocui.Gui) error {
				return s.renderStatsView(g)
			})
		}
	}
}

func (s *System) typingRoundLayout(g *gocui.Gui) error {
	return s.renderTypingView(g, passage, s.completedPassage)
}

func (s *System) handleTypingInputs(g *gocui.Gui, completed chan bool) {
	og := passage

outer:
	for {
		select {
		case t := <-s.tr.keyStroke:
			g.Update(func(g *gocui.Gui) error {
				v, _ := g.View("paragraph")
				return s.updatePromptView(v, og, t.key)
			})
		case <-completed:
			break outer
		}
	}

}

func (s *System) updatePromptView(v *gocui.View, og string, input gocui.Key) error {

	wordOutput := make([]string, len(s.tr.words))

	for i, w := range s.tr.words {
		wordOutput[i] = w.text
	}

	canMoveToNextWord := s.checkIfCanMoveOnToNextWord()

	if input == gocui.KeySpace {
		if canMoveToNextWord {
			s.tr.currentWordIndex++
			s.tr.nextWord <- true
			if s.tr.currentWordIndex == len(s.tr.words) {
				s.completedPassage <- true
				s.completedPassage <- true
				s.completedPassage <- true
				return nil
			}
		}
	}

	wordOutput[s.tr.currentWordIndex] = markWord(wordOutput[s.tr.currentWordIndex], s.isCorrectSoFar() || canMoveToNextWord)

	v.Clear()
	fmt.Fprintln(v, strings.Join(wordOutput, " "))
	return nil
}

func (s *System) typingEditor(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {

	if !s.tr.hasStarted {
		s.tr.hasStarted = true
		go func() {
			start := time.Now()
			s.tr.data = append(s.tr.data, 0)
		outer:
			for {
				select {
				case <-time.After(time.Second / 2):
					if s.tr.currentWordIndex < len(s.tr.words) {
						s.gui.Update(func(g *gocui.Gui) error {
							wv, _ := g.View("WPM")
							chars := s.tr.words[s.tr.currentWordIndex].index
							t := time.Now()
							curWPM := float64(chars) / t.Sub(start).Seconds() / 5 * 60
							s.tr.data = append(s.tr.data, curWPM)
							wv.Clear()
							fmt.Fprintln(wv, int(curWPM))
							return nil
						})
					}
				case <-s.completedPassage:
					break outer
				}
			}
		}()
	}

	ks := KeyStroke{
		key:  key,
		char: ch,
	}
	s.tr.keyStroke <- ks
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
			nw := <-s.tr.nextWord
			if nw {
				s.gui.Update(func(g *gocui.Gui) error {
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
