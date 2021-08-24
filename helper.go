package main

import (
	"fmt"
	"strings"

	"github.com/jroimartin/gocui"
)

func processPrompt(p string) []Word {
	pr := strings.Split(p, " ")
	lenSoFar := 0
	w := make([]Word, len(pr))
	for i := 0; i < len(pr); i++ {
		w[i].text = pr[i]
		w[i].index += lenSoFar
		lenSoFar += len(pr[i]) + 1
	}
	return w
}

func (s *System) checkIfCanMoveOnToNextWord() bool {
	return strings.TrimSpace(s.gui.CurrentView().ViewBuffer()) == s.words[s.currentWordIndex].text
}

func (s *System) isCorrectSoFar() bool {
	buffer := strings.TrimSpace(s.gui.CurrentView().ViewBuffer())
	return strings.HasPrefix(s.words[s.currentWordIndex].text, buffer) || buffer == ""
}

func (s *System) canMoveCursorForward(inputCorrect bool, currentIndex int, currentWordIndex int) bool {
	return inputCorrect && currentIndex < len(s.originalPrompt)-1 && strings.HasPrefix(s.words[currentWordIndex].text, strings.TrimSpace(s.gui.CurrentView().ViewBuffer()))
}

func markWord(word string, isCorrect bool) string {
	if isCorrect {
		return "\033[1;46m" + word + "\033[0m"
	} else {
		return "\033[37;1;41m" + word + "\033[0m"
	}
}

func printInitialPrompt(v *gocui.View, s string) {
	arr := strings.Split(s, " ")

	if len(arr) > 0 {
		arr[0] = markWord(arr[0], true)
		fmt.Fprintln(v, strings.Join(arr, " "))
	}
}

func isFinalWord() {

}
