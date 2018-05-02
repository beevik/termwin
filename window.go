package termwin

import termbox "github.com/nsf/termbox-go"

// A Window is an instance of a termwin control.
type Window interface {
	onKey(ev termbox.Event) error
	onDraw()
	getCursor() (x, y int, show bool)
}
