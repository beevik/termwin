package termwin

import (
	"errors"

	tb "github.com/nsf/termbox-go"
)

// EditBoxFlags define settings for an EditBox.
type EditBoxFlags byte

const (
	// EditBoxWordWrap causes the edit box to word-wrap a line of text when
	// its length reaches the right edge of the screen.
	EditBoxWordWrap EditBoxFlags = 1 << iota
)

// An EditBox represents a editable area of text on the screen. Many common text
// editing controls are allowed within an EditBox: cursor movement, character
// insertion and deleteion, text selection, copy/cut/paste, etc.
type EditBox struct {
	screenBox
	flags EditBoxFlags
}

// NewEditBox creates a new EditBox control with the specified screen
// position and size.
func NewEditBox(x, y, width, height int, flags EditBoxFlags) *EditBox {
	e := &EditBox{
		screenBox: newScreenBox(x, y, width, height),
		flags:     flags,
	}
	addWindow(e)
	return e
}

// getCursor returns the absolute screen position of the cursor.
func (e *EditBox) getCursor() (x, y int, show bool) {
	return e.screenBox.getCursor()
}

func (e *EditBox) onDraw() {
	e.Draw()
}

func (e *EditBox) onKey(ev tb.Event) error {
	e.modifiers = ev.Mod

	switch ev.Key {
	case tb.KeyArrowLeft, tb.KeyCtrlB:
		e.CursorLeft()
	case tb.KeyArrowRight, tb.KeyCtrlF:
		e.CursorRight()
	case tb.KeyArrowUp, tb.KeyCtrlP:
		e.CursorUp()
	case tb.KeyArrowDown, tb.KeyCtrlN:
		e.CursorDown()
	case tb.KeyHome:
		if (ev.Mod & tb.ModCtrl) != 0 {
			e.CursorStartOfBuffer()
		} else {
			e.CursorStartOfLine()
		}
	case tb.KeyCtrlA:
		e.CursorStartOfLine()
	case tb.KeyEnd:
		if (ev.Mod & tb.ModCtrl) != 0 {
			e.CursorEndOfBuffer()
		} else {
			e.CursorEndOfLine()
		}
	case tb.KeyCtrlE:
		e.CursorEndOfLine()
	case tb.KeyPgdn:
		e.CursorPageDown()
	case tb.KeyPgup:
		e.CursorPageUp()
	case tb.KeyDelete, tb.KeyCtrlD:
		e.DeleteChar()
	case tb.KeyBackspace, tb.KeyBackspace2:
		e.DeleteCharLeft()
	case tb.KeyCtrlC:
		e.CopyToClipboard()
	case tb.KeyCtrlV:
		e.PasteFromClipboard()
	case tb.KeyCtrlX:
		e.CutToClipboard()
	case tb.KeySpace:
		e.InsertChar(charSpace)
	case tb.KeyEnter:
		e.InsertChar(charNewline)
	default:
		if ev.Ch == '`' {
			return errors.New("exit")
		}
		if ev.Ch != 0 {
			e.InsertChar(ev.Ch)
		}
	}

	return nil
}
