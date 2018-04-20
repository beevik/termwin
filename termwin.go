package termwin

import (
	tb "github.com/nsf/termbox-go"
)

var c context

type context struct {
	windows []window
	focus   window
}

func addWindow(w window) {
	if len(c.windows) == 0 {
		c.focus = w
	}
	c.windows = append(c.windows, w)
}

// Init must be called before any termwin controls can be used.
func Init() error {
	err := tb.Init()
	if err != nil {
		return err
	}

	tb.SetInputMode(tb.InputEsc)
	return nil
}

// Flush flushes the contents of the back buffer to the screen display.
func Flush() {
	for _, w := range c.windows {
		w.onDraw()
	}

	if c.focus != nil {
		c.focus.onSetCursor()
	}

	tb.Flush()
}

// Close shuts down the termwin system.
func Close() {
	tb.Close()
}

// Poll polls the system for an input event
func Poll() {
	switch ev := tb.PollEvent(); ev.Type {
	case tb.EventKey:
		if c.focus != nil {
			c.focus.onKey(ev)
		}
	case tb.EventError:
		panic(ev.Err)
	}
}
