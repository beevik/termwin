package termwin

import (
	tb "github.com/nsf/termbox-go"
)

var c context

type context struct {
	windows []Window
	focus   Window
	kb      []byte
	escaped bool
}

func addWindow(w Window) {
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

	tb.SetInputMode(tb.InputAlt)
	return nil
}

// Close shuts down the termwin system.
func Close() {
	tb.Close()
}

// Size returns the current dimensions of the screen.
func Size() (x, y int) {
	return tb.Size()
}

// Flush flushes the contents of the back buffer to the screen display.
func Flush() {
	for _, w := range c.windows {
		w.onDraw()
	}

	if c.focus == nil {
		tb.HideCursor()
	} else {
		x, y, show := c.focus.getCursor()
		if show {
			tb.SetCursor(x, y)
		} else {
			tb.HideCursor()
		}
	}

	tb.Flush()
}

// SetFocus removes the cursor focus from any window it is currently on and
// adds focus to the specified window. If you pass nil for the window,
// SetFocus removes focus from all windows.
func SetFocus(w Window) {
	c.focus = w
}

// Poll polls the system for an input event
func Poll() error {
	switch ev := tb.PollEvent(); ev.Type {
	case tb.EventKey:
		Logf("Ch=0x%02X Key=0x%04X Mod=0x%02X\n", ev.Ch, ev.Key, ev.Mod)
		if ev.Ch != 0 {
			if c.escaped {
				c.kb = append(c.kb, byte(ev.Ch))
				if (ev.Ch >= 'a' && ev.Ch <= 'z') || (ev.Ch >= 'A' && ev.Ch <= 'Z') || ev.Ch == '~' {
					handleEscSeq(&ev)
					c.escaped, c.kb = false, c.kb[:0]
				} else {
					break
				}
			}
			if ev.Ch == '[' && (ev.Mod&tb.ModAlt) != 0 {
				c.escaped = true
				break
			}
		}

		if c.focus != nil {
			err := c.focus.onKey(ev)
			if err != nil {
				return err
			}
		}

	case tb.EventError:
		return ev.Err
	}

	return nil
}

type keymod struct {
	Key tb.Key
	Mod tb.Modifier
}

var escSeq = map[string]keymod{
	"1;2C": {tb.KeyArrowRight, tb.ModShift},
	"1;5C": {tb.KeyArrowRight, tb.ModCtrl},
	"1;2D": {tb.KeyArrowLeft, tb.ModShift},
	"1;5D": {tb.KeyArrowLeft, tb.ModCtrl},
	"1;2A": {tb.KeyArrowUp, tb.ModShift},
	"1;5A": {tb.KeyArrowUp, tb.ModCtrl},
	"1;2B": {tb.KeyArrowDown, tb.ModShift},
	"1;5B": {tb.KeyArrowDown, tb.ModCtrl},
	"1;2F": {tb.KeyEnd, tb.ModShift},
	"1;5F": {tb.KeyEnd, tb.ModCtrl},
	"1;6F": {tb.KeyEnd, tb.ModCtrl | tb.ModShift},
	"1;2H": {tb.KeyHome, tb.ModShift},
	"1;5H": {tb.KeyHome, tb.ModCtrl},
	"1;6H": {tb.KeyHome, tb.ModCtrl | tb.ModShift},
	"5;2~": {tb.KeyPgup, tb.ModShift},
	"5;5~": {tb.KeyPgup, tb.ModCtrl},
	"6;2~": {tb.KeyPgdn, tb.ModShift},
	"6;5~": {tb.KeyPgdn, tb.ModCtrl},
}

func handleEscSeq(ev *tb.Event) {
	s := escSeq[string(c.kb)]
	ev.Ch, ev.Key, ev.Mod = 0, s.Key, s.Mod
	//Logf("%s => %d,%d\n", string(c.kb), ev.Key, ev.Mod)
}
