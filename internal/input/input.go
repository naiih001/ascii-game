package input

import "github.com/gdamore/tcell/v2"

type Action int

const (
	ActionNone Action = iota
	ActionMoveUp
	ActionMoveDown
	ActionMoveLeft
	ActionMoveRight
	ActionQuit
	ActionResize
)

func Poll(screen tcell.Screen, out chan<- Action) {
	for {
		ev := screen.PollEvent()
		if ev == nil {
			continue
		}

		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch {
			case ev.Key() == tcell.KeyEscape || ev.Key() == tcell.KeyCtrlC:
				out <- ActionQuit
			case ev.Key() == tcell.KeyUp || ev.Rune() == 'k':
				out <- ActionMoveUp
			case ev.Key() == tcell.KeyDown || ev.Rune() == 'j':
				out <- ActionMoveDown
			case ev.Key() == tcell.KeyLeft || ev.Rune() == 'h':
				out <- ActionMoveLeft
			case ev.Key() == tcell.KeyRight || ev.Rune() == 'l':
				out <- ActionMoveRight
			}
		case *tcell.EventResize:
			screen.Sync()
			out <- ActionResize
		}
	}
}
