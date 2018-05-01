package termwin

import (
	tb "github.com/nsf/termbox-go"
)

var c context

type context struct {
	windows []window
	focus   window
	kb      []byte
	escaped bool
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

	tb.SetInputMode(tb.InputAlt)
	return nil
}

// Flush flushes the contents of the back buffer to the screen display.
func Flush() {
	for _, w := range c.windows {
		w.onDraw()
	}

	if c.focus != nil {
		x, y, _ := c.focus.getCursor()
		tb.SetCursor(x, y)
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
			c.focus.onKey(ev)
		}

	case tb.EventError:
		panic(ev.Err)
	}
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
	Logf("%s => %d,%d\n", string(c.kb), ev.Key, ev.Mod)
}
