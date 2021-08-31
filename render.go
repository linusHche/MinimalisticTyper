package main

import (
	"fmt"
	"log"

	"github.com/guptarohit/asciigraph"
	"github.com/jroimartin/gocui"
)

func (s *System) renderTypingView() error {
	maxX, maxY := s.g.Size()

	s.g.Cursor = true
	paragraphBottom := maxY - maxY/4

	if v, err := s.g.SetView("paragraph", 0, 0, maxX-1, paragraphBottom); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		printInitialPrompt(v, s.tr.originalPrompt)
		v.Wrap = true
		v.Title = "Prompt"
	}

	statsBottom := paragraphBottom + 1

	if v, err := s.g.SetView("WPM", maxX-10, statsBottom-4, maxX-2, paragraphBottom-1); err != nil {
		v.Title = "WPM"
	}

	if v, err := s.g.SetView("input", 0, statsBottom+1, maxX-1, statsBottom+4); err != nil {
		go s.handleTypingInputs()
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Editable = true
		v.Editor = gocui.EditorFunc(s.typingEditor)

		v.Wrap = true
		v.Title = "Type Here"
		s.g.SetCurrentView("input")
	}
	s.g.DeleteKeybinding("input", gocui.KeyCtrlL, gocui.ModNone)
	if err := s.g.SetKeybinding("input", gocui.KeyCtrlL, gocui.ModNone, s.nextTypingRound); err != nil {
		log.Panicln(err)
	}
	if v, err := s.g.SetView("skip", 1, statsBottom-3, maxX/4, paragraphBottom); err != nil {
		fmt.Fprintln(v, "Press CTRL + L to skip.")
		v.Frame = false
	}
	return nil
}

func (s *System) renderStatsView() error {
	viewName := "stats"
	maxX, maxY := s.g.Size()

	if v, err := s.g.SetView(viewName, 0, 0, maxX-1, maxY-1); err != nil {
		v.Title = "Stats"
		graph := asciigraph.Plot(s.tr.data, asciigraph.Height(maxY*90/100), asciigraph.Offset(2), asciigraph.Precision(0), asciigraph.Width(maxX*9/10))
		fmt.Fprintln(v, graph)
		s.g.SetCurrentView(viewName)
		s.g.DeleteKeybinding("stats", gocui.KeyEnter, gocui.ModNone)
		if err := s.g.SetKeybinding(viewName, gocui.KeyEnter, gocui.ModNone, s.nextTypingRound); err != nil {
			log.Panicln(err)
		}
	}

	if v, err := s.g.SetView("averageWPM", maxX-10, 1, maxX-2, 3); err != nil {
		v.Title = "WPM"
		averageWPM := 0.0
		for _, i := range s.tr.data {
			averageWPM += i
		}
		averageWPM /= float64(len(s.tr.data)) - 1
		fmt.Fprintln(v, int(averageWPM))
	}

	if v, err := s.g.SetView("continue", maxX/2, maxY-3, maxX-1, maxY-1); err != nil {
		fmt.Fprintln(v, "Press Enter to continue.")
	}

	return nil
}

func (s *System) nextTypingRound(g *gocui.Gui, v *gocui.View) error {
	s.completedPassage <- false
	s.completedPassage <- false
	s.completedPassage <- false
	s.newPassage <- getRandomPassage(s.passages)
	return nil
}
