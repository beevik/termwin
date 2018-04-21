package termwin

import (
	"unicode/utf8"

	tb "github.com/nsf/termbox-go"
)

// EditBoxFlags define settings for an EditBox.
type EditBoxFlags byte

const (
	// EditBoxWordWrap causes the edit box to word-wrap a line of text when
	// its length reaches the right edge of the screen.
	EditBoxWordWrap EditBoxFlags = 1 << iota

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
	cells []tb.Cell
	flags byte
}

func newRow(width int) row {
	return row{
		cells: make([]tb.Cell, 0, width),
		flags: 0,
	}
}

func (r *row) grow(n int) {
	l := len(r.cells) + n
	if l > cap(r.cells) {
		cells := make([]tb.Cell, l, max(l, cap(r.cells)*2))
		copy(cells, r.cells)
	}
	r.cells = append(r.cells, tb.Cell{})
}

// An EditBox represents a editable text control with fixed screen dimensions.
type EditBox struct {
	flags        EditBoxFlags
	size         vec2  // dimensions of the edit box
	screenCorner vec2  // screen coordinate of top-left corner
	viewRect     rect  // buffer currently visible
	dirtyRect    rect  // portions of the buffer that have been updated
	rows         []row // all rows in the buffer
	cursor       vec2  // current cursor position
}

func (e *EditBox) onDraw() {
	e.Draw()
}

func (e *EditBox) onKey(ev tb.Event) {
	switch ev.Key {
	case tb.KeyEsc:
		Logln(e.Contents())
	case tb.KeyArrowLeft, tb.KeyCtrlB:
		e.CursorLeft()
	case tb.KeyArrowRight, tb.KeyCtrlF:
		e.CursorRight()
	case tb.KeyArrowUp:
		e.CursorUp()
	case tb.KeyArrowDown:
		e.CursorDown()
	case tb.KeyPgdn:
		e.SetCursor(e.cursor.x, e.cursor.y+e.size.y)
	case tb.KeyPgup:
		e.SetCursor(e.cursor.x, max(e.cursor.y-e.size.y, 0))
	case tb.KeyDelete, tb.KeyCtrlD:
		e.DeleteChar()
	case tb.KeyBackspace, tb.KeyBackspace2:
		if e.cursor.x > 0 || e.cursor.y > 0 {
			e.CursorLeft()
			e.DeleteChar()
		}
	case tb.KeySpace:
		e.InsertChar(' ')
	case tb.KeyHome, tb.KeyCtrlA:
		e.updateCursor(0, e.cursor.y)
	case tb.KeyEnd, tb.KeyCtrlE:
		e.updateCursor(e.EndOfRow(e.cursor.y), e.cursor.y)
	case tb.KeyEnter:
		e.InsertChar('\n')
	default:
		if ev.Ch == '`' {
			panic("exit")
		}
		if ev.Ch != 0 {
			e.InsertChar(ev.Ch)
		}
	}
}

func (e *EditBox) onSetCursor() {
	x := e.cursor.x - e.viewRect.x0 + e.screenCorner.x
	y := e.cursor.y - e.viewRect.y0 + e.screenCorner.y
	tb.SetCursor(x, y)
}

// NewEditBox creates a new EditBox control with the specified screen
// position and size.
func NewEditBox(x, y, width, height int, flags EditBoxFlags) *EditBox {
	e := &EditBox{
		flags:        flags,
		size:         vec2{width, height},
		screenCorner: vec2{x, y},
		viewRect:     newRect(0, 0, width, height),
		dirtyRect:    rect{0, 0, maxValue, maxValue},
		rows:         []row{newRow(width)},
	}
	addWindow(e)
	return e
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
			cr := &e.rows[cy]
			nr := &e.rows[cy+1]
			nr.cells = cr.cells[cx:]
			cr.cells = cr.cells[:cx]
			e.updateDirtyRect(rect{cx, cy, maxValue, cy + 1})

		case charLinefeed:
			e.updateCursor(0, -1)
		}

	default:
		cr := &e.rows[cy]
		cr.grow(1)
		if cx <= len(cr.cells) {
			copy(cr.cells[cx+1:], cr.cells[cx:])
		}
		cr.cells[cx] = tb.Cell{Ch: ch}
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

// CursorLeft moves the cursor left, shifting to the end of the previous line
// if the cursor is at column 0.
func (e *EditBox) CursorLeft() {
	if e.cursor.x > 0 {
		e.adjustCursor(-1, 0)
	} else if e.cursor.y > 0 {
		y := e.cursor.y - 1
		r := &e.rows[y]
		x := len(r.cells)
		e.updateCursor(x, y)
	}
}

// CursorRight moves the cursor right, shifting to the next line if the cursor
// is at the right-most column of the current line.
func (e *EditBox) CursorRight() {
	r := &e.rows[e.cursor.y]
	if e.cursor.x < len(r.cells) {
		e.adjustCursor(+1, 0)
	} else if e.cursor.y+1 < len(e.rows) {
		e.updateCursor(0, e.cursor.y+1)
	}
}

// CursorDown moves the cursor down a line.
func (e *EditBox) CursorDown() {
	if e.cursor.y+1 < len(e.rows) {
		x, y := e.cursor.x, e.cursor.y+1
		r := &e.rows[y]
		if x > len(r.cells) {
			x = len(r.cells)
		}
		e.updateCursor(x, y)
	}
}

// CursorUp moves the cursor up a line.
func (e *EditBox) CursorUp() {
	if e.cursor.y > 0 {
		x, y := e.cursor.x, e.cursor.y-1
		r := &e.rows[y]
		if x > len(r.cells) {
			x = len(r.cells)
		}
		e.updateCursor(x, y)
	}
}

// SetCursor sets the position of the cursor within the view buffer. Negative
// values position the cursor relative to the last column and row of the
// buffer. A value of -1 for x indicates the end of the row. A value of -1
// for y indicates the last row.
func (e *EditBox) SetCursor(x, y int) {
	if y < 0 {
		y = len(e.rows) + y
		if y < 0 {
			y = 0
		}
	} else if y >= len(e.rows) {
		y = len(e.rows) - 1
	}

	cr := &e.rows[y]

	if x < 0 {
		x = len(e.rows[y].cells) + 1 + x
		if x < 0 {
			x = 0
		}
	} else if x > len(cr.cells) {
		x = len(cr.cells)
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

	encodeRow := func(i int) {
		r := &e.rows[i]
		for _, c := range r.cells {
			sz := utf8.EncodeRune(rbuf, c.Ch)
			buf = append(buf, rbuf[:sz]...)
		}
	}

	i := 0
	for n := len(e.rows) - 1; i < n; i++ {
		encodeRow(i)
		buf = append(buf, '\n')
	}
	encodeRow(i)

	return string(buf)
}

// Draw updates the contents of the EditBox on the screen.
func (e *EditBox) Draw() {
	buf := tb.CellBuffer()
	stride, _ := tb.Size()

	r := intersection(e.dirtyRect, e.viewRect)
	width, height := r.x1-r.x0, r.y1-r.y0

	boffset := e.screenCorner.x + (e.screenCorner.y+r.y0-e.viewRect.y0)*stride

	ymax := min(r.y1, len(e.rows))
	ymin := min(max(r.y0, 0), ymax)
	for y := ymin; y < ymax; y++ {
		row := &e.rows[y]
		xmax := min(r.x1, len(row.cells))
		xmin := min(max(r.x0, 0), xmax)
		o := boffset + r.x0 - e.viewRect.x0
		copy(buf[o:], row.cells[xmin:xmax])
		clearCells(buf[o+xmax-xmin : o+width])
		boffset += stride
	}

	remain := height - (ymax - ymin)
	for y := 0; y < remain; y++ {
		clearCells(buf[boffset : boffset+width])
		boffset += stride
	}

	e.dirtyRect = emptyRect
}

var emptyCell = tb.Cell{Ch: ' '}

func clearCells(c []tb.Cell) {
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
		e.updateDirtyRect(e.viewRect)
	case e.cursor.x < e.viewRect.x0:
		dx := e.viewRect.x0 - e.cursor.x
		e.viewRect.x0 -= dx
		e.viewRect.x1 -= dx
		e.updateDirtyRect(e.viewRect)
	}

	switch {
	case e.cursor.y >= e.viewRect.y1:
		dy := e.cursor.y - e.viewRect.y1 + 1
		e.viewRect.y0 += dy
		e.viewRect.y1 += dy
		e.updateDirtyRect(e.viewRect)
	case e.cursor.y < e.viewRect.y0:
		dy := e.viewRect.y0 - e.cursor.y
		e.viewRect.y0 -= dy
		e.viewRect.y1 -= dy
		e.updateDirtyRect(e.viewRect)
	}
}
