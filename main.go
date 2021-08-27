package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/guptarohit/asciigraph"
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
}

var passage string = "Born too late to explore the Earth. Born too soon to explore the universe. Born just in time to code in Golang."

// var passage string = "Born too"

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	s := System{
		newPassage:       make(chan string),
		completedPassage: make(chan bool),
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

func (s *System) renderTypingView(g *gocui.Gui, p string, completed chan bool) error {
	maxX, maxY := g.Size()

	g.Cursor = true
	paragraphBottom := maxY - maxY/2

	if v, err := g.SetView("paragraph", 0, 0, maxX-1, paragraphBottom); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		printInitialPrompt(v, passage)
		v.Wrap = true
		v.Title = "Prompt"
	}

	statsBottom := paragraphBottom + 1

	if v, err := g.SetView("WPM", maxX-10, statsBottom-4, maxX-2, paragraphBottom-1); err != nil {
		v.Title = "WPM"
	}

	if v, err := g.SetView("input", 0, statsBottom+1, maxX-1, statsBottom+4); err != nil {
		go s.handleType(g, completed)
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Editable = true
		v.Editor = gocui.EditorFunc(s.typingEditor)

		v.Wrap = true
		v.Title = "Type Here"
		g.SetCurrentView("input")
	}
	return nil
}

func (s *System) renderStatsView(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	if v, err := g.SetView("stats", 0, 0, maxX-1, maxY-1); err != nil {
		v.Title = "stats"
		graph := asciigraph.Plot(s.tr.data, asciigraph.Height(maxY*90/100), asciigraph.Offset(2), asciigraph.Precision(0), asciigraph.Width(maxX*9/10))
		fmt.Fprintln(v, graph)
		g.SetCurrentView("stats")
	}

	return nil
}

func (s *System) handleViews(g *gocui.Gui) {
	for {
		select {
		case p := <-s.newPassage:
			g.Update(func(g *gocui.Gui) error {
				g.DeleteView("stats")
				return s.renderTypingView(g, p, s.completedPassage)
			})
		case <-s.completedPassage:
			g.Update(func(g *gocui.Gui) error {
				// g.DeleteView("paragraph")
				// g.DeleteView("WPM")
				// g.DeleteView("input")
				return s.renderStatsView(g)
			})
		}
	}
}

func (s *System) typingRoundLayout(g *gocui.Gui) error {
	return s.renderTypingView(g, passage, s.completedPassage)
}

func (s *System) handleType(g *gocui.Gui, completed chan bool) {
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
		default:
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
					s.gui.Update(func(g *gocui.Gui) error {
						wv, _ := g.View("WPM")
						if s.tr.currentWordIndex < len(s.tr.words) {
							chars := s.tr.words[s.tr.currentWordIndex].index
							t := time.Now()
							curWPM := float64(chars) / t.Sub(start).Seconds() / 5 * 60
							s.tr.data = append(s.tr.data, curWPM)
							wv.Clear()
							fmt.Fprintln(wv, int(curWPM))
						}
						return nil
					})
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
