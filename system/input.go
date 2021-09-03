package system

import (
	"fmt"
	"time"

	"github.com/jroimartin/gocui"
)

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
					t := time.Now()

					chars := tr.words[tr.currentWordIndex].index
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
