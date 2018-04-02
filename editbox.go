package termwin

import (
	termbox "github.com/nsf/termbox-go"
)

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
	contPrev byte = 1 << iota // continues previous line (word-wrap)
	contNext                  // continues on next line (word-wrap)
)

const (
	charNewline   rune = '\n'
	charBackspace rune = '\b'
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

// InsertChar inserts a new character at the current cursor position and
// advances the cursor by one column.
func (e *EditBox) InsertChar(ch rune) {
}

// InsertString inserts an entire string at the current cursor position
// and advances the cursor by the length of the string.
func (e *EditBox) InsertString(s string) {
}

// InsertRow inserts a new row at the current cursor position, leaving
// the cursor position unchanged.
func (e *EditBox) InsertRow() {
}

// DeleteChar deletes a single character at the current cursor position.
func (e *EditBox) DeleteChar() {
}

// DeleteChars deletes multiple characters starting from the current cursor
// position.
func (e *EditBox) DeleteChars(n int) {
}

// DeleteRow deletes the entire row containing the cursor.
func (e *EditBox) DeleteRow() {
}

// LastRow returns the row number of the last row in the view buffer.
func (e *EditBox) LastRow() int {
	return len(e.viewRows) - 1
}

// EndOfRow returns the column position representing the end of row `y`. Pass
// a value of -1 for `y` to find the end of the row containing the cursor.
// If the requested row doesn't exist, this returns -1.
func (e *EditBox) EndOfRow(y int) int {
	if y == -1 {
		y = e.cursor.y
	}
	if y >= len(e.viewRows) {
		return -1
	}
	row := &e.viewRows[y]
	return len(row.cells)
}

// Size returns the width and height of the EditBox on screen.
func (e *EditBox) Size() (width, height int) {
	return e.screenRect.x1 - e.screenRect.x0, e.screenRect.y1 - e.screenRect.y0
}

// SetCursor sets the position of the cursor within the view buffer. Negative
// values position the cursor relative to the last column and row of the
// buffer. A value of -1 for x or y represents the cursor's current column or
// row number, respectively.
func (e *EditBox) SetCursor(x, y int) {
}

// Cursor returns the cursor's current column and row within the view buffer.
func (e *EditBox) Cursor() (x, y int) {
	return e.cursor.x, e.cursor.y
}

// SetView adjusts the buffer position currently representing the top-left
// corner of the visible EditBox.
func (e *EditBox) SetView(x, y int) {
}

// View returns the buffer position currently representing the top-left
// corner of the visible EditBox.
func (e *EditBox) View() (x, y int) {
	return e.viewRect.x0, e.viewRect.y0
}

// Contents returns the entire contents of the EditBox buffer.
func (e *EditBox) Contents() string {
	return ""
}
