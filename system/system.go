package system

import (
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
	isTypingRound    bool
}

func MakeSystem(g *gocui.Gui, ps []string) *System {
	return &System{
		newPassage:       make(chan string),
		completedPassage: make(chan bool, 3),
		g:                g,
		tr:               makeTypingRound(getRandomPassage(ps)),
		passages:         ps,
		isTypingRound:    true,
	}
}

func (s *System) HandleViews(g *gocui.Gui) {
	for {
		select {
		case p := <-s.newPassage:
			s.isTypingRound = true
			s.restartTypingView(p)
		case c := <-s.completedPassage:
			if c {
				s.isTypingRound = false
			}
		}
	}
}

func (s *System) Layout(g *gocui.Gui) error {
	if s.isTypingRound {
		return s.renderTypingView()
	}
	return s.renderStatsView()
}

func makeTypingRound(p string) *TypingRound {
	return &TypingRound{
		nextWord:         make(chan bool),
		keyStroke:        make(chan KeyStroke),
		originalPrompt:   p,
		words:            processPrompt(p),
		currentWordIndex: 0,
		inputCorrect:     false,
		hasStarted:       false,
		data:             make([]float64, 0),
	}
}

func (s *System) nextTypingRound(g *gocui.Gui, v *gocui.View) error {
	s.completedPassage <- false
	s.completedPassage <- false
	s.completedPassage <- false
	s.newPassage <- getRandomPassage(s.passages)
	return nil
}
