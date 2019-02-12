package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/japanoise/termbox-util"
	termbox "github.com/nsf/termbox-go"
)

// MaxWidth - max width of a line in rune-widths
const MaxWidth int = 80

type line struct {
	Data       string
	width      int
	saved      bool
	savedbytes int
}

type state struct {
	Display  []*line
	sy       int
	filename string
	file     *os.File
	word     string
	wc       int
}

// Note: sy will be uninitialised, so make sure to set it up w/
// termbox afterwards
func createState(filename string) (*state, error) {
	ret := state{}

	ret.filename = filename

	ret.Display = []*line{&line{"", 0, false, 0}}

	var err error
	ret.file, err = os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

func (s *state) recalcDisplay() {
	termbox.Sync()
	_, s.sy = termbox.Size()

	if len(s.Display) > s.sy-2 {
		s.Display = s.Display[s.sy/2:]
	}
}

func (s *state) dump() {
	f, err := os.Create("dump.json")
	if err != nil {
		return
	}
	defer f.Close()

	b, err := json.Marshal(s)
	if err != nil {
		return
	}
	f.Write(b)
}

func (s *state) draw() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	for i, line := range s.Display {
		termutil.Printstring(line.Data, 0, 1+i)
	}
	ld := len(s.Display)
	lline := s.Display[ld-1]
	if lline.Data == "" {
		termutil.Printstring(s.word, 0, ld)
		termbox.SetCursor(termutil.RunewidthStr(s.word), ld)
	} else {
		termutil.Printstring(s.word, lline.width+1, ld)
		termbox.SetCursor(lline.width+1+termutil.RunewidthStr(s.word), ld)
	}
	termutil.PrintstringColored(termbox.AttrReverse, fmt.Sprintf("%s -- %d words added", s.filename, s.wc), 0, 0)
	termbox.Flush()
}

func (s *state) addWord() {
	if s.word == "" {
		return
	}
	defer func() { s.word = ""; s.wc++; s.recalcDisplay() }()
	w := termutil.RunewidthStr(s.word)
	if len(s.Display) == 0 {
		s.Display = []*line{&line{s.word, w, false, 0}}
		return
	}
	lline := s.Display[len(s.Display)-1]
	if lline.width == 0 {
		lline.Data = s.word
		lline.width = w
		return
	}
	if lline.width+w+1 > MaxWidth {
		if lline.saved && len(lline.Data) > lline.savedbytes {
			s.file.WriteString(lline.Data[lline.savedbytes:])
		} else {
			s.file.WriteString(lline.Data)
		}
		s.file.WriteString("\n")
		s.Display = append(s.Display, &line{s.word, w, false, 0})
		return
	}
	lline.Data += " " + s.word
	lline.width += w + 1
}

func (s *state) newline() {
	s.addWord()
	lline := s.Display[len(s.Display)-1]
	if lline.saved && len(lline.Data) > lline.savedbytes {
		s.file.WriteString(lline.Data[lline.savedbytes:])
	} else {
		s.file.WriteString(lline.Data)
	}
	s.file.WriteString("\n")
	s.Display = append(s.Display, &line{"", 0, false, 0})
	s.recalcDisplay()
}

func (s *state) save() error {
	s.addWord()
	lline := s.Display[len(s.Display)-1]
	if lline.saved && len(lline.Data) > lline.savedbytes {
		s.file.WriteString(lline.Data[lline.savedbytes:])
		lline.savedbytes = len(lline.Data)
	} else {
		lline.saved = true
		s.file.WriteString(lline.Data)
		lline.savedbytes = len(lline.Data)
	}
	return s.file.Sync()
}

func (s *state) exit() error {
	s.save()
	lline := s.Display[len(s.Display)-1]
	if lline.Data != "" {
		s.file.WriteString("\n")
	}
	return s.file.Close()
}
