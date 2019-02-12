package main

import (
	"fmt"
	"os"
	"path/filepath"
	"unicode/utf8"

	"github.com/japanoise/termbox-util"
	"github.com/nsf/termbox-go"
)

func errdie(err string) {
	fmt.Fprintf(os.Stderr, "%s: %s\n", filepath.Base(os.Args[0]), err)
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		errdie("please provide a filename")
	}

	var err error
	s, err := createState(os.Args[1])
	if err != nil {
		errdie(err.Error())
	}
	defer s.exit()

	err = termbox.Init()
	if err != nil {
		errdie(err.Error())
	}
	defer termbox.Close()
	s.recalcDisplay()

	for {
		s.draw()
		ev := termbox.PollEvent()
		switch ev.Type {
		case termbox.EventResize:
			s.recalcDisplay()
		case termbox.EventKey:
			key := termutil.ParseTermboxEvent(ev)
			switch key {
			case "C-q", "C-x":
				return
			case "DEL", "C-h":
				s.word = ""
			case "C-s":
				s.save()
			case "C-d":
				s.dump()
			case " ":
				s.addWord()
			case "RET":
				s.newline()
			default:
				if utf8.RuneCountInString(key) == 1 {
					s.word += key
				}
			}
		}
	}
}
