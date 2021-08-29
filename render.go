package main

import (
	"fmt"
	"log"
	"math/rand"

	"github.com/guptarohit/asciigraph"
	"github.com/jroimartin/gocui"
)

func (s *System) renderTypingView(g *gocui.Gui, p string, completed chan bool) error {
	maxX, maxY := g.Size()

	g.Cursor = true
	paragraphBottom := maxY - maxY/2

	if v, err := g.SetView("paragraph", 0, 0, maxX-1, paragraphBottom); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		printInitialPrompt(v, p)
		v.Wrap = true
		v.Title = "Prompt"
	}

	statsBottom := paragraphBottom + 1

	if v, err := g.SetView("WPM", maxX-10, statsBottom-4, maxX-2, paragraphBottom-1); err != nil {
		v.Title = "WPM"
	}

	if v, err := g.SetView("input", 0, statsBottom+1, maxX-1, statsBottom+4); err != nil {
		go s.handleTypingInputs(g, completed)
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
	viewName := "stats"
	maxX, maxY := g.Size()

	if v, err := g.SetView(viewName, 0, 0, maxX-1, maxY-1); err != nil {
		v.Title = "Stats"
		graph := asciigraph.Plot(s.tr.data, asciigraph.Height(maxY*90/100), asciigraph.Offset(2), asciigraph.Precision(0), asciigraph.Width(maxX*9/10))
		fmt.Fprintln(v, graph)
		g.SetCurrentView(viewName)
		if err := g.SetKeybinding(viewName, gocui.KeyEnter, gocui.ModNone, s.quitStatsView); err != nil {
			log.Panicln(err)
		}
	}

	if v, err := g.SetView("averageWPM", maxX-10, 1, maxX-2, 3); err != nil {
		v.Title = "WPM"
		averageWPM := 0.0
		for _, i := range s.tr.data {
			averageWPM += i
		}
		averageWPM /= float64(len(s.tr.data)) - 1
		fmt.Fprintln(v, int(averageWPM))
	}

	if v, err := g.SetView("continue", maxX/2, maxY-3, maxX-1, maxY-1); err != nil {
		fmt.Fprintln(v, "Press Enter to continue.")
	}

	return nil
}

func (s *System) quitStatsView(g *gocui.Gui, v *gocui.View) error {
	i := rand.Intn(len(s.passages))
	s.newPassage <- s.passages[i]
	return nil
}
