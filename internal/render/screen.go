package render

import (
	"errors"
	"os"

	"github.com/gdamore/tcell/v2"
	"golang.org/x/term"
)

type Screen struct {
	screen tcell.Screen
}

func New() (*Screen, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		return nil, errors.New("game requires an interactive terminal (TTY) on stdin/stdout")
	}

	s, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}

	if err := s.Init(); err != nil {
		return nil, err
	}

	s.SetStyle(tcell.StyleDefault)

	return &Screen{screen: s}, nil
}

func (s *Screen) Raw() tcell.Screen {
	return s.screen
}

func (s *Screen) Clear() {
	s.screen.Clear()
}

func (s *Screen) Size() (int, int) {
	return s.screen.Size()
}

func (s *Screen) DrawRune(x, y int, ch rune, style tcell.Style) {
	s.screen.SetContent(x, y, ch, nil, style)
}

func (s *Screen) Show() {
	s.screen.Show()
}

func (s *Screen) Close() {
	s.screen.Fini()
}
