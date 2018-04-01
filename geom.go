package termwin

type vec2 struct {
	x, y int
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
