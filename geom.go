package termwin

const (
	minValue = -(1 << 31)
	maxValue = ((1 << 31) - 1)
)

var emptyRect = rect{maxValue, maxValue, minValue, minValue}

type coord struct {
	x, y int
}

func (c coord) equals(c2 coord) bool {
	return c.y == c2.y && c.x == c2.x
}

func (c coord) lessThan(c2 coord) bool {
	switch {
	case c.y < c2.y:
		return true
	case c.y > c2.y:
		return false
	default:
		return c.x < c2.x
	}
}

func (c coord) lessThanOrEqual(c2 coord) bool {
	switch {
	case c.y < c2.y:
		return true
	case c.y > c2.y:
		return false
	default:
		return c.x <= c2.x
	}
}

func (c coord) greaterThan(c2 coord) bool {
	switch {
	case c.y > c2.y:
		return true
	case c.y < c2.y:
		return false
	default:
		return c.x > c2.x
	}
}

func (c coord) greaterThanOrEqual(c2 coord) bool {
	switch {
	case c.y > c2.y:
		return true
	case c.y < c2.y:
		return false
	default:
		return c.x >= c2.x
	}
}

func (c coord) inRange(r crange) bool {
	return c.lessThan(r.c1) && c.greaterThanOrEqual(r.c0)
}

type crange struct {
	c0 coord
	c1 coord
}

func (r crange) empty() bool {
	return r.c0.equals(r.c1)
}

func (r crange) ordered() crange {
	if r.c0.greaterThan(r.c1) {
		return crange{r.c1, r.c0}
	}
	return r
}

type rect struct {
	x0, y0 int
	x1, y1 int
}

func newRect(x, y, width, height int) rect {
	return rect{x, y, x + width, y + height}
}

func (r rect) empty() bool {
	return r.x1 <= r.x0 || r.y1 <= r.y0
}

func union(r1, r2 rect) rect {
	return rect{
		x0: min(r1.x0, r2.x0),
		y0: min(r1.y0, r2.y0),
		x1: max(r1.x1, r2.x1),
		y1: max(r1.y1, r2.y1),
	}
}

func intersection(r1, r2 rect) rect {
	return rect{
		x0: max(r1.x0, r2.x0),
		y0: max(r1.y0, r2.y0),
		x1: min(r1.x1, r2.x1),
		y1: min(r1.y1, r2.y1),
	}
}

func intersects(r1, r2 rect) bool {
	return !intersection(r1, r2).empty()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
