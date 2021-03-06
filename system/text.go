package system

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/jroimartin/gocui"
)

func markWord(word string, isCorrect bool) string {
	if isCorrect {
		return "\033[1;30;47m" + word + "\033[0m"
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

func getRandomPassage(p []string) string {
	rand.Seed(time.Now().UnixNano())
	i := rand.Intn(len(p))
	return p[i]
}

func (s *System) checkIfCanMoveOnToNextWord() bool {
	return strings.TrimSpace(s.g.CurrentView().ViewBuffer()) == s.tr.words[s.tr.currentWordIndex].text
}

func (s *System) isCorrectSoFar() bool {
	buffer := strings.TrimSpace(s.g.CurrentView().ViewBuffer())
	return strings.HasPrefix(s.tr.words[s.tr.currentWordIndex].text, buffer) || buffer == ""
}

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
