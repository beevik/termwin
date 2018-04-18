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
	charLinefeed  rune = '\r'
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
		dirtyRect:  rect{0, 0, maxValue, maxValue},
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
	cx, cy := e.cursor.x, e.cursor.y
	switch {
	case ch < 32:
		switch ch {
		case charBackspace:
			e.updateCursor(max(cx-1, 0), -1)

		case charNewline:
			e.updateCursor(0, cy+1)
			e.InsertRow()

		case charLinefeed:
			e.updateCursor(0, -1)
		}

	default:
		cr := &e.rows[cy]
		cr.grow(1)
		if cx <= len(cr.cells) {
			copy(cr.cells[cx+1:], cr.cells[cx:])
		}
		cr.cells[cx] = termbox.Cell{Ch: ch}
		e.updateDirtyRect(rect{cx, cy, maxValue, cy + 1})
		e.adjustCursor(+1, 0)
	}
}

// InsertString inserts an entire string at the current cursor position
// and advances the cursor by the length of the string.
func (e *EditBox) InsertString(s string) {
	for _, ch := range s {
		e.InsertChar(ch)
	}
}

// InsertRow inserts a new row at the current cursor position. The cursor
// moves to the beginning of the inserted row.
func (e *EditBox) InsertRow() {
	cr := e.cursor.y
	e.rows = append(e.rows, row{})
	copy(e.rows[cr+1:], e.rows[cr:])
	e.rows[cr] = newRow(e.size.x)
	e.updateCursor(0, -1)
	e.updateDirtyRect(rect{0, cr, maxValue, maxValue})
}

// DeleteChar deletes a single character at the current cursor position.
func (e *EditBox) DeleteChar() {
	cx, cy := e.cursor.x, e.cursor.y
	cr := &e.rows[cy]

	// At end of line? Merge lines.
	if e.cursor.x >= len(cr.cells) {
		if cy+1 < len(e.rows) {
			nr := &e.rows[cy+1]
			cr.cells = append(cr.cells, nr.cells...)
			e.rows = append(e.rows[:cy+1], e.rows[cy+2:]...)
			e.updateDirtyRect(rect{0, cy, maxValue, maxValue})
		}
		return
	}

	// Remove character from line
	copy(cr.cells[cx:], cr.cells[cx+1:])
	cr.cells = cr.cells[:len(cr.cells)-1]
	e.updateDirtyRect(rect{cx, cy, maxValue, cy + 1})
}

// DeleteChars deletes multiple characters starting from the current cursor
// position.
func (e *EditBox) DeleteChars(n int) {
	for i := 0; i < n; i++ {
		e.DeleteChar()
	}
}

// DeleteRow deletes the entire row containing the cursor.
func (e *EditBox) DeleteRow() {
	cy := e.cursor.y
	if cy+1 < len(e.rows) {
		e.rows = append(e.rows[:cy+1], e.rows[cy+2:]...)
		e.updateDirtyRect(rect{0, cy, maxValue, maxValue})
	} else {
		e.rows = e.rows[:cy+1]
		e.updateDirtyRect(rect{0, cy, maxValue, maxValue})
	}
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
	e.updateCursor(x, y)
}

// Cursor returns the cursor's current column and row within the view buffer.
func (e *EditBox) Cursor() (x, y int) {
	return e.cursor.x, e.cursor.y
}

// SetView adjusts the buffer position currently representing the top-left
// corner of the visible EditBox.
func (e *EditBox) SetView(x, y int) {
	e.viewRect = rect{x, y, x + e.size.x, y + e.size.y}
	e.updateDirtyRect(e.viewRect)
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
	e.dirtyRect = emptyRect

	pitch, height := r.x1-r.x0, r.y1-r.y0

	ymax := min(r.y1, len(e.rows))
	ymin := min(max(r.y0, 0), ymax)
	for y := ymin; y < ymax; y++ {
		row := &e.rows[y]
		xmax := min(r.x1, len(row.cells))
		xmin := min(max(r.x0, 0), xmax)
		copy(dst[offset:], row.cells[xmin:xmax])
		clearCells(dst[offset+xmax-xmin : offset+pitch])
		offset += cx
	}

	remain := height - (ymax - ymin)
	for y := 0; y < remain; y++ {
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

func (e *EditBox) adjustCursor(dx, dy int) {
	e.cursor.x += dx
	e.cursor.y += dy
	e.updateView()
}

func (e *EditBox) updateCursor(x, y int) {
	if x != -1 {
		e.cursor.x = x
	}
	if y != -1 {
		e.cursor.y = y
	}
	e.updateView()
}

func (e *EditBox) updateView() {
	switch {
	case e.cursor.x >= e.viewRect.x1:
		dx := e.cursor.x - e.viewRect.x1 + 1
		e.viewRect.x0 += dx
		e.viewRect.x1 += dx
	case e.cursor.x < e.viewRect.x0:
		dx := e.viewRect.x0 - e.cursor.x
		e.viewRect.x0 -= dx
		e.viewRect.x1 -= dx
	}
	switch {
	case e.cursor.y >= e.viewRect.y1:
		dy := e.cursor.y - e.viewRect.y1 + 1
		e.viewRect.y0 += dy
		e.viewRect.y1 += dy
	case e.cursor.y < e.viewRect.y0:
		dy := e.viewRect.y0 - e.cursor.y
		e.viewRect.y0 -= dy
		e.viewRect.y1 -= dy
	}
}
