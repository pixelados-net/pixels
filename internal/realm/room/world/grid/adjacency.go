package grid

// Adjacent reports whether two points share an edge or corner without being equal.
func Adjacent(first Point, second Point) bool {
	dx := absolute(int(first.X) - int(second.X))
	dy := absolute(int(first.Y) - int(second.Y))
	return (dx != 0 || dy != 0) && dx <= 1 && dy <= 1
}

// absolute returns a non-negative integer magnitude.
func absolute(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
