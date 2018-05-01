package termwin

import termbox "github.com/nsf/termbox-go"

type window interface {
	onKey(ev termbox.Event)
	onDraw()
	getCursor() (x, y int, show bool)
}
