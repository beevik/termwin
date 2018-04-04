package termwin

import (
	"unicode/utf8"

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

func (r *row) grow(n int) {
	l := len(r.cells) + n
	if l > cap(r.cells) {
		cells := make([]termbox.Cell, l, cap(r.cells)*2)
		copy(cells, r.cells)
	}
	r.cells = append(r.cells, termbox.Cell{})
}

// An EditBox represents a editable text control with fixed screen dimensions.
type EditBox struct {
	mode       EditBoxMode
	size       vec2  // dimensions of the edit box
	screenRect rect  // screen dimensions of the edit box
	viewRect   rect  // buffer currently visible
	dirtyRect  rect  // portions of the buffer that have been updated
	rows       []row // all rows in the buffer
	cursor     vec2  // current cursor position
}

// NewEditBox creates a new EditBox control with the specified screen
// position and size.
func NewEditBox(x, y, width, height int, mode EditBoxMode) *EditBox {
	return &EditBox{
		mode:       mode,
		size:       vec2{width, height},
		screenRect: newRect(x, y, width, height),
		viewRect:   newRect(0, 0, width, height),
		rows:       []row{newRow(width)},
	}
}

// Write the contents of a UTF8-formatted buffer starting at the current
// cursor position. This function allows you to use standard formatted
// output functions like `fmt.Fprintf` with an EditBox control.
func (e *EditBox) Write(p []byte) (n int, err error) {
	for len(p) > 0 {
		ch, sz := utf8.DecodeRune(p)
		if err != nil {
			return 0, err
		}
		p = p[sz:]
		e.InsertChar(ch)
	}
	return 0, nil
}

// InsertChar inserts a new character at the current cursor position and
// advances the cursor by one column.
func (e *EditBox) InsertChar(ch rune) {
	switch {
	case ch < 32:
		switch ch {
		case '\b':
			e.cursor.x = max(e.cursor.x-1, 0)

		case '\n':
			e.cursor.y++
			e.InsertRow()
			fallthrough
		case '\r':
			e.cursor.x = 0
		}

	default:
		cr := &e.rows[e.cursor.y]
		cr.grow(1)
		cx := e.cursor.x
		if cx <= len(cr.cells) {
			copy(cr.cells[cx+1:], cr.cells[cx:])
		}
		cr.cells[cx] = termbox.Cell{Ch: ch}
		e.updateDirtyRect(rect{e.cursor.x, e.cursor.y, maxValue, e.cursor.y + 1})
		e.cursor.x++
	}
}

// InsertString inserts an entire string at the current cursor position
// and advances the cursor by the length of the string.
func (e *EditBox) InsertString(s string) {
	for _, ch := range s {
		e.InsertChar(ch)
	}
}

// InsertRow inserts a new row at the current cursor position, leaving
// the cursor position unchanged.
func (e *EditBox) InsertRow() {
	cr := e.cursor.y
	e.rows = append(e.rows, row{})
	copy(e.rows[cr+1:], e.rows[cr:])
	e.rows[cr] = newRow(e.size.x)
	e.updateDirtyRect(rect{0, e.cursor.y, maxValue, maxValue})
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
	return len(e.rows) - 1
}

// EndOfRow returns the column position representing the end of row `y`. Pass
// a value of -1 for `y` to find the end of the row containing the cursor.
// If the requested row doesn't exist, this returns -1.
func (e *EditBox) EndOfRow(y int) int {
	if y == -1 {
		y = e.cursor.y
	}
	if y >= len(e.rows) {
		return -1
	}
	row := &e.rows[y]
	return len(row.cells)
}

// Size returns the width and height of the EditBox on screen.
func (e *EditBox) Size() (width, height int) {
	return e.size.x, e.size.y
}

// SetCursor sets the position of the cursor within the view buffer. Negative
// values position the cursor relative to the last column and row of the
// buffer. A value of -1 for x or y represents the cursor's current column or
// row number, respectively.
func (e *EditBox) SetCursor(x, y int) {
	if y < 0 {
		y = len(e.rows) + y
		if y < 0 {
			y = 0
		}
	}
	if x < 0 {
		x = len(e.rows[y].cells) + x
		if x < 0 {
			x = 0
		}
	}
	e.cursor = vec2{x, y}
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
	var rbuf = make([]byte, 4)
	var buf []byte
	for i, n := 0, len(e.rows); i < n; i++ {
		r := &e.rows[i]
		for _, c := range r.cells {
			utf8.EncodeRune(rbuf, c.Ch)
			buf = append(buf, rbuf...)
		}
		if i+1 < n {
			buf = append(buf, '\n')
		}
	}
	return string(buf)
}

// Draw updates the contents of the EditBox on the screen.
func (e *EditBox) Draw() {
	dst := termbox.CellBuffer()
	cx, _ := termbox.Size()
	offset := e.screenRect.x0 + e.screenRect.y0*cx

	r := intersection(e.dirtyRect, e.viewRect)
	pitch := r.x1 - r.x0
	e.dirtyRect = emptyRect

	ym := min(r.y1, len(e.rows))
	for y := r.y0; y < ym; y++ {
		row := &e.rows[y]
		xm := min(r.x1, len(row.cells))
		copy(dst[offset:], row.cells[r.x0:xm])
		clearCells(dst[offset+xm-r.x0 : offset+pitch])
		offset += cx
	}
	for y := ym; y < r.y1; y++ {
		clearCells(dst[offset : offset+pitch])
		offset += cx
	}
}

var emptyCell = termbox.Cell{Ch: ' '}

func clearCells(c []termbox.Cell) {
	for i := range c {
		c[i] = emptyCell
	}
}

func (e *EditBox) updateDirtyRect(r rect) {
	e.dirtyRect = union(e.dirtyRect, r)
}
