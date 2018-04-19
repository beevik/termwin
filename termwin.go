package termwin

import termbox "github.com/nsf/termbox-go"

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
	err := termbox.Init()
	if err != nil {
		return err
	}

	termbox.SetInputMode(termbox.InputEsc | termbox.InputMouse)
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

	termbox.Flush()
}

// Close shuts down the termwin system.
func Close() {
	termbox.Close()
}

// Poll polls the system for an input event
func Poll() {
	switch ev := termbox.PollEvent(); ev.Type {
	case termbox.EventKey:
		if c.focus != nil {
			c.focus.onKey(ev)
		}
	case termbox.EventError:
		panic(ev.Err)
	}
}
