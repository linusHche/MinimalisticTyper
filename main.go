package main

import (
	"MinimalisticTyper/system"
	"fmt"
	"log"
	"os"

	"github.com/jroimartin/gocui"
)

var file, _ = os.OpenFile("logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)

func main() {

	ps, err1 := processGivenPrompt()
	if err1 != nil {
		fmt.Println(err1.Error())
		os.Exit(1)
	}

	g, err2 := gocui.NewGui(gocui.OutputNormal)
	if err2 != nil {
		log.Panicln(err2)
	}

	defer g.Close()

	log.SetOutput(file)

	s := system.MakeSystem(g, ps)
	g.SetManagerFunc(s.Layout)
	go s.HandleViews(g)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
