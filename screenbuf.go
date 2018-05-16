package termwin

import (
	"unicode/utf8"

	tb "github.com/nsf/termbox-go"
)

const (
	charSpace     = ' '
	charNewline   = '\n'
	charLinefeed  = '\r'
	charBackspace = '\b'
)

var (
	emptyCell = tb.Cell{Ch: charSpace}
)

// A row represents a single line of text within the screen buffer.
type row struct {
	cells []tb.Cell // edit buffer cells in this row
}

func newRow(width int) row {
	return row{
		cells: make([]tb.Cell, 0, width),
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

// A screenBox represents a rectangle of text that can be displayed on the
// console at a given location.
type screenBox struct {
	size      coord       // screen dimensions of the buffer
	corner    coord       // screen coordinate of top-left corner
	view      rect        // visible portion of the buffer
	dirty     rect        // portion of the buffer that needs an update
	rows      []row       // all rows in the edit buffer
	cursor    coord       // current cursor position
	lastX     int         // cursor X position after last horz move
	modifiers tb.Modifier // modifier keys currently down
	selecting bool        // cursor in selecting mode
	selection crange      // current selection range
}

// newScreenBox creates a new EditBox control with the specified screen
// position and size.
func newScreenBox(x, y, width, height int) screenBox {
	return screenBox{
		size:   coord{width, height},
		corner: coord{x, y},
		view:   newRect(0, 0, width, height),
		dirty:  rect{0, 0, maxValue, maxValue},
		rows:   []row{newRow(width)},
	}
}

// getCursor returns the absolute screen position of the cursor.
func (b *screenBox) getCursor() (x, y int, show bool) {
	x = b.cursor.x - b.view.x0 + b.corner.x
	y = b.cursor.y - b.view.y0 + b.corner.y
	show = true
	return
}

// Write the contents of a UTF8-formatted buffer starting at the current
// cursor position. This function allows you to use standard formatted
// output functions like `fmt.Fprintf` with an EditBox control.
func (b *screenBox) Write(p []byte) (n int, err error) {
	for len(p) > 0 {
		ch, sz := utf8.DecodeRune(p)
		if err != nil {
			return 0, err
		}
		p = p[sz:]
		b.InsertChar(ch)
	}
	return 0, nil
}

// InsertChar inserts a new character at the current cursor position and
// advances the cursor by one column.
func (b *screenBox) InsertChar(ch rune) {
	if b.selecting {
		b.deleteRange(b.selection.ordered())
		b.selecting = false
	}

	cx, cy := b.cursor.x, b.cursor.y
	switch {
	case ch < 32:
		switch ch {
		case charNewline:
			b.updateCursor(0, cy+1)
			b.InsertRow()
			currRow := &b.rows[cy]
			nextRow := &b.rows[cy+1]
			nextRow.cells = append(nextRow.cells, currRow.cells[cx:]...)
			currRow.cells = append(currRow.cells[:cx], emptyCell)
			b.updateDirtyRect(rect{cx, cy, maxValue, cy + 1})

		case charLinefeed:
			b.updateCursor(0, cy)
		}

	default:
		row := &b.rows[cy]
		row.grow(1)
		copy(row.cells[cx+1:], row.cells[cx:])
		row.cells[cx] = tb.Cell{Ch: ch}
		b.updateDirtyRect(rect{cx, cy, maxValue, cy + 1})
		b.updateCursor(cx+1, cy)
	}

	b.lastX = b.cursor.x
}

// InsertString inserts an entire string at the current cursor position
// and advances the cursor by the length of the string.
func (b *screenBox) InsertString(s string) {
	for _, ch := range s {
		b.InsertChar(ch)
	}
}

// InsertRow inserts a new row at the current cursor position. The cursor
// moves to the beginning of the inserted row.
func (b *screenBox) InsertRow() {
	cy := b.cursor.y
	b.rows = append(b.rows, row{})
	copy(b.rows[cy+1:], b.rows[cy:])
	b.rows[cy] = newRow(b.size.x)
	b.updateCursor(0, cy)
	b.updateDirtyRect(rect{0, cy, maxValue, maxValue})
}

// DeleteChar deletes a single character at the current cursor position.
func (b *screenBox) DeleteChar() {
	if b.selecting {
		b.deleteRange(b.selection.ordered())
		b.selecting = false
		return
	}

	cx, cy := b.cursor.x, b.cursor.y
	rl := b.rowLen(cy)
	row := &b.rows[cy]

	// At end of line? Merge lines.
	if cx >= rl {
		if cy+1 < len(b.rows) {
			nr := &b.rows[cy+1]
			row.cells = append(row.cells[:rl], nr.cells...)
			b.rows = append(b.rows[:cy+1], b.rows[cy+2:]...)
			b.updateDirtyRect(rect{0, cy, maxValue, maxValue})
		}
		return
	}

	// Remove character from line
	copy(row.cells[cx:], row.cells[cx+1:])
	row.cells = row.cells[:len(row.cells)-1]
	b.updateDirtyRect(rect{cx, cy, maxValue, cy + 1})
}

// DeleteCharLeft deletes the character to the left of the cursor and moves
// the cursor to the position of the deleted character. If the cursor is at
// the start of the line, the newline is removed.
func (b *screenBox) DeleteCharLeft() {
	if b.cursor.x == 0 && b.cursor.y == 0 {
		return
	}

	if b.selecting {
		b.deleteRange(b.selection.ordered())
		b.selecting = false
		return
	}

	b.CursorLeft()
	b.DeleteChar()
}

// DeleteChars deletes multiple characters starting from the current cursor
// position.
func (b *screenBox) DeleteChars(n int) {
	b.deleteChars(n, b.cursor.x, b.cursor.y)
}

// DeleteRow deletes the entire row containing the cursor.
func (b *screenBox) DeleteRow() {
	cy := b.cursor.y
	if cy+1 < len(b.rows) {
		b.rows = append(b.rows[:cy+1], b.rows[cy+2:]...)
		b.updateDirtyRect(rect{0, cy, maxValue, maxValue})
	} else {
		b.rows = b.rows[:cy+1]
		b.updateDirtyRect(rect{0, cy, maxValue, maxValue})
	}
}

// LastRow returns the row number of the last row in the buffer.
func (b *screenBox) LastRow() int {
	return len(b.rows) - 1
}

// Size returns the width and height of the EditBox on screen.
func (b *screenBox) Size() (width, height int) {
	return b.size.x, b.size.y
}

// CopyToClipboard copies the current selection to the clipboard.
func (b *screenBox) CopyToClipboard() {
	if b.selecting {
		ClipboardSet(b.getRange(b.selection.ordered()))
	} else {
		ClipboardClear()
	}
}

// CutToClipboard copies the current selection to the clipboard and then
// deletes it from the edit buffer.
func (b *screenBox) CutToClipboard() {
	if b.selecting {
		r := b.selection.ordered()
		ClipboardSet(b.getRange(r))
		b.deleteRange(r)
		b.selecting = false
	}
}

// PasteFromClipboard pastes the current clipboard contents to the edit buffer
// at the current cursor position.
func (b *screenBox) PasteFromClipboard() {
	if b.selecting {
		b.deleteRange(b.selection)
		b.selecting = false
	}

	s := ClipboardGet()
	if s != "" {
		b.InsertString(s)
	}
}

// Selection returns the contents of the substring currently selected in the
// edit buffer.
func (b *screenBox) Selection() string {
	if b.selecting {
		return b.getRange(b.selection.ordered())
	}
	return ""
}

// Cursor returns the cursor's current column and row within the edit buffer.
func (b *screenBox) Cursor() (x, y int) {
	return b.cursor.x, b.cursor.y
}

// CursorSet sets the position of the cursor within the edit buffer. Negative
// values position the cursor relative to the last column and row of the
// buffer. A value of -1 for x indicates the end of the row. A value of -1
// for y indicates the last row.
func (b *screenBox) CursorSet(x, y int) {
	if y < 0 {
		y = len(b.rows) + y
		if y < 0 {
			y = 0
		}
	} else if y >= len(b.rows) {
		y = len(b.rows) - 1
	}

	rl := b.rowLen(y)
	if x < 0 {
		x = rl + 1 + x
		if x < 0 {
			x = 0
		}
	} else if x > rl {
		x = rl
	}

	b.updateCursor(x, y)
	b.lastX = b.cursor.x
}

// CursorLeft moves the cursor left, shifting to the end of the previous line
// if the cursor is at column 0.
func (b *screenBox) CursorLeft() {
	cx, cy := b.cursor.x, b.cursor.y
	if cx > 0 {
		b.updateCursor(cx-1, cy)
	} else if cy > 0 {
		cx := b.rowLen(cy - 1)
		b.updateCursor(cx, cy-1)
	}
	b.lastX = b.cursor.x
}

// CursorRight moves the cursor right, shifting to the next line if the cursor
// is at the right-most column of the current line.
func (b *screenBox) CursorRight() {
	cx, cy := b.cursor.x, b.cursor.y
	rl := b.rowLen(cy)
	if cx < rl {
		b.updateCursor(cx+1, cy)
	} else if cy+1 < len(b.rows) {
		b.updateCursor(0, cy+1)
	}
	b.lastX = b.cursor.x
}

// CursorDown moves the cursor down a line.
func (b *screenBox) CursorDown() {
	if b.cursor.y+1 >= len(b.rows) {
		return
	}

	cx, cy := b.lastX, b.cursor.y+1
	rl := b.rowLen(cy)
	if cx > rl {
		cx = rl
	}
	b.updateCursor(cx, cy)
}

// CursorUp moves the cursor up a line.
func (b *screenBox) CursorUp() {
	if b.cursor.y == 0 {
		return
	}

	cx, cy := b.lastX, b.cursor.y-1
	rl := b.rowLen(cy)
	if cx > rl {
		cx = rl
	}
	b.updateCursor(cx, cy)
}

// CursorWordStart moves the cursor to the start of the word.
func (b *screenBox) CursorWordStart() {
	b.CursorLeft()
}

// CursorWordEnd moves the cursor to end of the word.
func (b *screenBox) CursorWordEnd() {
	c := b.cursor
	for b.isValid(c) && !isCellChar(b.getCell(c)) {
		c = b.nextCell(c)
	}
	for b.isValid(c) && isCellChar(b.getCell(c)) {
		c = b.nextCell(c)
	}
	b.updateCursor(c.x, c.y)
}

// CursorStartOfBuffer moves the cursor to the start of the edit buffer.
func (b *screenBox) CursorStartOfBuffer() {
	b.updateCursor(0, 0)
	b.lastX = b.cursor.x
}

// CursorStartOfLine moves the cursor to the start of the current line.
func (b *screenBox) CursorStartOfLine() {
	b.updateCursor(0, b.cursor.y)
	b.lastX = b.cursor.x
}

// CursorEndOfBuffer moves the cursor to the end of the edit buffer.
func (b *screenBox) CursorEndOfBuffer() {
	cy := len(b.rows) - 1
	cx := b.rowLen(cy)
	b.updateCursor(cx, cy)
	b.lastX = b.cursor.x
}

// CursorEndOfLine moves the cursor to the end of the current line.
func (b *screenBox) CursorEndOfLine() {
	cy := b.cursor.y
	cx := b.rowLen(cy)
	b.updateCursor(cx, cy)
	b.lastX = b.cursor.x
}

// CursorPageDown moves the cursor down a page.
func (b *screenBox) CursorPageDown() {
	cx, cy := b.lastX, b.cursor.y
	cy += b.size.y - 1
	if cy >= len(b.rows) {
		cy = len(b.rows) - 1
	}

	rl := b.rowLen(cy)
	if cx > rl {
		cx = rl
	}

	b.updateCursor(cx, cy)
}

// CursorPageUp moves the cursor up a page.
func (b *screenBox) CursorPageUp() {
	cx, cy := b.lastX, b.cursor.y
	cy -= b.size.y - 1
	if cy < 0 {
		cy = 0
	}

	rl := b.rowLen(cy)
	if cx > rl {
		cx = rl
	}

	b.updateCursor(cx, cy)
}

// View returns the buffer position currently representing the top-left
// corner of the visible EditBox.
func (b *screenBox) View() (x, y int) {
	return b.view.x0, b.view.y0
}

// SetView adjusts the buffer position currently representing the top-left
// corner of the visible EditBox.
func (b *screenBox) SetView(x, y int) {
	b.view = rect{x, y, x + b.size.x, y + b.size.y}
	b.updateDirtyRect(b.view)
}

// Contents returns the entire contents of the edit buffer.
func (b *screenBox) Contents() string {
	r := crange{
		c0: coord{0, 0},
		c1: coord{maxValue, len(b.rows) - 1},
	}
	return b.getRange(r)
}

// getCells returns cells in columns x0 through x1 on row y. This function
// assumes y is a valid row and x0 <= x1.
func (b *screenBox) getCells(y, x0, x1 int) []tb.Cell {
	r := &b.rows[y]
	rl := len(r.cells)
	if y+1 < len(b.rows) {
		rl--
	}

	x0, x1 = min(x0, rl), min(x1, rl)
	return r.cells[x0:x1]
}

// appendCellChars appends the characters in a cell slice to a slice of bytes
// and returns the updated slice.
func appendCellChars(buf []byte, c []tb.Cell) []byte {
	var rbuf [4]byte
	for _, cc := range c {
		sz := utf8.EncodeRune(rbuf[:], cc.Ch)
		buf = append(buf, rbuf[:sz]...)
	}
	return buf
}

// getRange returns the contents of the edit buffer covering the specified
// range. This function assumes the y values in the range are valid.
func (b *screenBox) getRange(r crange) string {
	var buf []byte

	x, y := r.c0.x, r.c0.y
	for ; y < r.c1.y; y++ {
		buf = appendCellChars(buf, b.getCells(y, x, maxValue))
		buf = append(buf, '\n')
		x = 0
	}
	buf = appendCellChars(buf, b.getCells(y, x, r.c1.x))

	return string(buf)
}

// deleteRange removes a range of text from the edit buffer.
func (b *screenBox) deleteRange(r crange) {
	x, y := r.c0.x, r.c0.y
	if y == r.c1.y {
		b.deleteCells(y, r.c0.x, r.c1.x)
	} else {
		for i, n := 0, r.c1.y-r.c0.y; i < n; i++ {
			b.deleteCells(y, x, maxValue)
		}
		b.deleteCells(y, x, x+r.c1.x)
	}
}

// deleteChars deletes up to n characters starting at position (x,y). The
// position (x,y) is assumed to be valid.
func (b *screenBox) deleteChars(n, x, y int) {
	for n > 0 {
		r := &b.rows[y]
		if len(r.cells) == 0 {
			break
		}

		nn := min(len(r.cells)-x, n)
		b.deleteCells(y, x, x+nn)
		n -= nn

		x = 0
	}
}

// deleteCells deletes cells in columns [x0:x1] on row y. This function
// assumes y is a valid row and x0 <= x1.
func (b *screenBox) deleteCells(y, x0, x1 int) {
	r := &b.rows[y]

	// fix bounds
	rl := len(r.cells)
	x0 = min(x0, rl)
	x1 = max(0, min(x1, rl))

	// delete cells
	b.updateDirtyRect(rect{x0, y, maxValue, y + 1})
	if x1 < rl || y+1 == len(b.rows) {
		r.cells = append(r.cells[:x0], r.cells[x1:]...)
	} else {
		nr := &b.rows[y+1]
		r.cells = append(r.cells[:x0], nr.cells...)
		b.rows = append(b.rows[:y+1], b.rows[y+2:]...)
		b.updateDirtyRect(rect{0, y + 1, maxValue, maxValue})
	}

	// adjust cursor
	switch {
	case b.cursor.y < y:
		// do nothing
	case b.cursor.y > y:
		b.cursor.y--
	case b.cursor.x >= x1:
		b.cursor.x -= (x1 - x0)
		b.lastX = b.cursor.x
	case b.cursor.x >= x0:
		b.cursor.x = x0
		b.lastX = b.cursor.x
	}
	b.updateView()
}

// Draw updates the contents of the EditBox on the screen.
func (b *screenBox) Draw() {
	buf := tb.CellBuffer()
	stride, _ := tb.Size()

	r := intersection(b.dirty, b.view)
	width, height := r.x1-r.x0, r.y1-r.y0

	boffset := b.corner.x + (b.corner.y+r.y0-b.view.y0)*stride

	ymax := min(r.y1, len(b.rows))
	ymin := min(max(r.y0, 0), ymax)
	for y := ymin; y < ymax; y++ {
		row := &b.rows[y]
		xmax := min(r.x1, len(row.cells))
		xmin := min(max(r.x0, 0), xmax)
		o := boffset + r.x0 - b.view.x0
		copy(buf[o:], row.cells[xmin:xmax])
		clearCells(buf[o+xmax-xmin : o+width])
		boffset += stride
	}

	remain := height - (ymax - ymin)
	for y := 0; y < remain; y++ {
		clearCells(buf[boffset : boffset+width])
		boffset += stride
	}

	b.dirty = emptyRect
}

// rowLen returns the length of a row, not including any terminating
// newline character.
func (b *screenBox) rowLen(y int) int {
	row := &b.rows[y]
	rl := len(row.cells)
	if rl > 0 && y < len(b.rows)-1 {
		return rl - 1
	}
	return rl
}

// getCell returns a copy of the cell at the requested coordinate.
func (b *screenBox) getCell(c coord) *tb.Cell {
	row := &b.rows[c.y]
	return &row.cells[c.x]
}

// cellAtPos returns a pointer to a back-cuffer cell at the requested
// position.
func (b *screenBox) cellAtPos(x, y int) *tb.Cell {
	if x < 0 || y < 0 || y >= len(b.rows) {
		return nil
	}
	row := &b.rows[y]
	rl := b.rowLen(y)
	if x >= rl {
		return nil
	}

	return &row.cells[x]
}

func (b *screenBox) nextCell(c coord) coord {
	rl := b.rowLen(c.y)
	if c.x < rl {
		return coord{c.x + 1, c.y}
	} else if c.y+1 < len(b.rows) {
		return coord{0, c.y + 1}
	} else {
		return c
	}
}

func (b *screenBox) isValid(c coord) bool {
	switch {
	case c.y >= len(b.rows):
		return false
	case c.x > len(b.rows[c.y].cells):
		return false
	default:
		return true
	}
}

// updateDirtyRect adds a rectangle to the currently dirty rectangle. The
// dirty rectangle is used to update the screen's backbuffer the next time it
// is drawn.
func (b *screenBox) updateDirtyRect(r rect) {
	b.dirty = union(b.dirty, r)
}

// updateCursor updates the position of the cursor. The new position is not
// validated.
func (b *screenBox) updateCursor(cx, cy int) {
	shiftDown := (b.modifiers & tb.ModShift) != 0
	switch {
	case shiftDown && !b.selecting:
		b.selection.c0 = b.cursor
		b.selection.c1 = b.cursor
		b.selecting = true
	case !shiftDown && b.selecting:
		b.unhighlight(b.selection.ordered())
		b.selecting = false
	}

	if b.selecting {
		b.updateSelection(cx, cy)
	}

	b.cursor.x, b.cursor.y = cx, cy
	b.updateView()
}

// updateSelection updates the currently selected range of text in the edit
// buffer.
func (b *screenBox) updateSelection(x, y int) {
	curr := coord{x, y}
	switch {
	case b.selection.c0.lessThan(b.selection.c1):
		switch {
		case curr.lessThan(b.selection.c0):
			b.unhighlight(crange{b.selection.c0, b.selection.c1})
			b.highlight(crange{curr, b.selection.c0})
		case curr.greaterThan(b.selection.c1):
			b.highlight(crange{b.selection.c1, curr})
		default:
			b.unhighlight(crange{curr, b.selection.c1})
		}

	default:
		switch {
		case curr.lessThan(b.selection.c1):
			b.highlight(crange{curr, b.selection.c1})
		case curr.greaterThan(b.selection.c0):
			b.unhighlight(crange{b.selection.c1, b.selection.c0})
			b.highlight(crange{b.selection.c0, curr})
		default:
			b.unhighlight(crange{b.selection.c1, curr})
		}
	}
	b.selection.c1 = curr
}

func (b *screenBox) highlight(r crange) {
	b.setCellAttribRange(r, tb.ColorBlack, tb.ColorWhite)
}

func (b *screenBox) unhighlight(r crange) {
	b.setCellAttribRange(r, tb.ColorDefault, tb.ColorDefault)
}

// setCellAttribRange adjusts the attributes of all cells within a range.
func (b *screenBox) setCellAttribRange(r crange, fg, bg tb.Attribute) {
	x, y := r.c0.x, r.c0.y
	for ; y < r.c1.y; y++ {
		row := &b.rows[y]
		for ; x < len(row.cells); x++ {
			setCellAttrib(&row.cells[x], fg, bg)
		}
		x = 0
	}
	for ; x < r.c1.x; x++ {
		setCellAttrib(b.cellAtPos(x, y), fg, bg)
	}

	b.updateDirtyRect(rect{0, r.c0.y, maxValue, r.c1.y + 1})
}

// updateView uses the current cursor position to make sure the text under
// the cursor is visible.
func (b *screenBox) updateView() {
	switch {
	case b.cursor.x >= b.view.x1:
		dx := b.cursor.x - b.view.x1 + 1
		b.view.x0 += dx
		b.view.x1 += dx
		b.updateDirtyRect(b.view)
	case b.cursor.x < b.view.x0:
		dx := b.view.x0 - b.cursor.x
		b.view.x0 -= dx
		b.view.x1 -= dx
		b.updateDirtyRect(b.view)
	}

	switch {
	case b.cursor.y >= b.view.y1:
		dy := b.cursor.y - b.view.y1 + 1
		b.view.y0 += dy
		b.view.y1 += dy
		b.updateDirtyRect(b.view)
	case b.cursor.y < b.view.y0:
		dy := b.view.y0 - b.cursor.y
		b.view.y0 -= dy
		b.view.y1 -= dy
		b.updateDirtyRect(b.view)
	}
}

func clearCells(c []tb.Cell) {
	for i := range c {
		c[i] = emptyCell
	}
}

func setCellAttrib(c *tb.Cell, fg, bg tb.Attribute) {
	c.Fg, c.Bg = fg, bg
}

func isCellChar(cell *tb.Cell) bool {
	return (cell.Ch >= 'a' && cell.Ch <= 'z') ||
		(cell.Ch >= 'A' && cell.Ch <= 'Z') ||
		(cell.Ch >= '0' && cell.Ch <= '9')
}
