package termwin

import termbox "github.com/nsf/termbox-go"

// EditBoxMode defines settings for an EditBox.
type EditBoxMode byte

const (
	// EditBoxWordWrap causes the edit box to word-wrap a line of text when
	// its length reaches the right edge of the screen.
	EditBoxWordWrap EditBoxMode = 1 << iota

	// EditBoxSingleRow allows only a single row of text. Carriage returns
	// and word-wrap are ignored.
	EditBoxSingleRow
)

const (
	contPrev byte = 1 << iota
	contNext
)

type row struct {
	cells []termbox.Cell
	flags byte
}

func newRow(width int) row {
	return row{
		cells: make([]termbox.Cell, 0, width),
		flags: 0,
	}
}

// An EditBox represents a editable text control with fixed screen dimensions.
type EditBox struct {
	screenRect rect
	viewRect   rect
	dirtyRect  rect
	viewRows   []row
	mode       EditBoxMode
	cursor     vec2
}

// NewEditBox creates a new EditBox control with the specified screen
// position and size.
func NewEditBox(x, y, width, height int, mode EditBoxMode) *EditBox {
	return &EditBox{
		screenRect: newRect(x, y, width, height),
		viewRect:   newRect(0, 0, width, height),
		viewRows:   []row{newRow(width)},
		mode:       mode,
	}
}

// InsertRune inserts a new rune at the current cursor position and advances
// the cursor.
func (e *EditBox) InsertRune(r rune) {
}

// InsertString inserts an entire string at the current cursor position
// and advances the cursor.
func (e *EditBox) InsertString(s string) {
}

// DeleteRune deletes a rune at the current cursor position.
func (e *EditBox) DeleteRune() {
}

// DeleteRunes deletes multiple runes starting from the current cursor
// position.
func (e *EditBox) DeleteRunes(n int) {
}

// RowCount returns the number of rows in the EditBox's buffer.
func (e *EditBox) RowCount() int {
	return len(e.viewRows)
}

// SetCursor sets the position of the cursor. Negative values position
// the cursor relative to the last column and row of the buffer.
func (e *EditBox) SetCursor(x, y int) {
}

// Cursor returns the current position of the cursor.
func (e *EditBox) Cursor() (x, y int) {
	return e.cursor.x, e.cursor.y
}

// SetTopLeft sets the buffer position to use as the top-left corner of
// the visible EditBox.
func (e *EditBox) SetTopLeft(x, y int) {
}

// TopLeft returns the buffer position currently representing the top-left
// corner of the visible EditBox.
func (e *EditBox) TopLeft() (x, y int) {
	return e.viewRect.x0, e.viewRect.y0
}

// Contents returns the entire contents of the EditBox buffer.
func (e *EditBox) Contents() string {
	return ""
}
