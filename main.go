package main

import (
	"fmt"
	"log"
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

type System struct {
	nextWord         chan bool
	keyStroke        chan KeyStroke
	originalPrompt   string
	words            []Word
	gui              *gocui.Gui
	currentWordIndex int
	inputCorrect     bool
	hasStarted       bool
	characters       int
}

var passage string = "Born too late to explore the Earth. Born too soon to explore the universe. Born just in time to code in Golang."

func main() {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func layout(g *gocui.Gui) error {
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

	// if v, err := g.SetView("graph", 0, paragraphBottom+1, maxX-1, statsBottom); err != nil {
	// 	if err != gocui.ErrUnknownView {
	// 		return err
	// 	}
	// 	v.Title = "WPM"
	// }

	if v, err := g.SetView("WPM", maxX-10, statsBottom-4, maxX-2, paragraphBottom-1); err != nil {
		fmt.Fprintln(v, 0)
		v.Title = "WPM"
	}

	if v, err := g.SetView("input", 0, statsBottom+1, maxX-1, statsBottom+4); err != nil {
		p := passage
		s := System{
			nextWord:         make(chan bool),
			keyStroke:        make(chan KeyStroke),
			originalPrompt:   p,
			words:            processPrompt(p),
			gui:              g,
			currentWordIndex: 0,
			inputCorrect:     false,
			hasStarted:       false,
			characters:       0,
		}
		go s.handleType(g)
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

func (s *System) handleType(g *gocui.Gui) {
	og := passage

	for {
		t := <-s.keyStroke
		g.Update(func(g *gocui.Gui) error {
			v, _ := g.View("paragraph")
			return s.updatePromptView(v, og, t.key)
		})
	}

}

func (s *System) updatePromptView(v *gocui.View, og string, input gocui.Key) error {

	wordOutput := make([]string, len(s.words))

	for i, w := range s.words {
		wordOutput[i] = w.text
	}

	canMoveToNextWord := s.checkIfCanMoveOnToNextWord()

	if input == gocui.KeySpace {
		if canMoveToNextWord {
			s.nextWord <- true
			s.currentWordIndex++
			if s.currentWordIndex == len(s.words) {
				return gocui.ErrQuit
			}
		}
	}

	wordOutput[s.currentWordIndex] = markWord(wordOutput[s.currentWordIndex], s.isCorrectSoFar() || canMoveToNextWord)

	v.Clear()
	fmt.Fprintln(v, strings.Join(wordOutput, " "))
	return nil
}

func (s *System) typingEditor(v *gocui.View, key gocui.Key, ch rune, mod gocui.Modifier) {

	if !s.hasStarted {
		s.hasStarted = true
		go func() {
			start := time.Now()
			var data []float64
			data = append(data, 0)
			for {
				<-time.After(time.Second / 2)
				s.gui.Update(func(g *gocui.Gui) error {
					// gv, _ := g.View("graph")
					wv, _ := g.View("WPM")
					if s.currentWordIndex < len(s.words) {
						chars := s.words[s.currentWordIndex].index
						t := time.Now()
						curWPM := float64(chars) / t.Sub(start).Seconds() / 5 * 60
						data = append(data, curWPM)
						// gv.Clear()
						wv.Clear()
						// width, height := gv.Size()
						// graph := asciigraph.Plot(data, asciigraph.Height(height*9/10), asciigraph.Offset(2), asciigraph.Precision(0), asciigraph.Width(width*9/10))
						// fmt.Fprintln(gv, graph)
						fmt.Fprintln(wv, int(curWPM))
						// fmt.Fprintln(sv, int(float64(chars)/t.Sub(start).Seconds()/5*60))
					}
					return nil
				})
			}
		}()
	}

	ks := KeyStroke{
		key:  key,
		char: ch,
	}
	s.keyStroke <- ks
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
			nw := <-s.nextWord
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
