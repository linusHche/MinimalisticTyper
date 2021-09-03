package system

import (
	"fmt"
	"strings"

	"github.com/jroimartin/gocui"
)

func (s *System) receivingTypingInputs() {

outer:
	for {
		select {
		case t := <-s.tr.keyStroke:
			s.updatePromptView(t.key)
		case <-s.completedPassage:
			break outer
		}
	}

}

func (s *System) updatePromptView(input gocui.Key) {
	s.g.Update(func(g *gocui.Gui) error {
		v, _ := g.View("paragraph")
		tr := s.tr
		wordOutput := make([]string, len(tr.words))

		for i, w := range tr.words {
			wordOutput[i] = w.text
		}

		canMoveToNextWord := s.checkIfCanMoveOnToNextWord()

		if input == gocui.KeySpace && canMoveToNextWord {
			tr.currentWordIndex++
			tr.nextWord <- true
			if tr.currentWordIndex == len(tr.words) {
				s.completedPassage <- true
				s.completedPassage <- true
				s.completedPassage <- true
				return nil
			}
		}

		wordOutput[tr.currentWordIndex] = markWord(wordOutput[tr.currentWordIndex], s.isCorrectSoFar() || canMoveToNextWord)

		v.Clear()
		fmt.Fprintln(v, strings.Join(wordOutput, " "))
		return nil
	})
}
