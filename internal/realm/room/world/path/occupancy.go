package path

// Occupancy stores blocked path positions.
type Occupancy struct {
	// blocked stores occupied nodes by path node key.
	blocked map[nodeKey]struct{}
}

// NewOccupancy creates occupancy from blocked positions.
func NewOccupancy(positions []Position) Occupancy {
	occupancy := Occupancy{blocked: make(map[nodeKey]struct{}, len(positions))}
	for _, position := range positions {
		occupancy.blocked[nodeFromPosition(position)] = struct{}{}
	}

	return occupancy
}

// Empty reports whether occupancy has no blocked positions.
func (occupancy Occupancy) Empty() bool {
	return len(occupancy.blocked) == 0
}

// Len returns the number of occupied positions.
func (occupancy Occupancy) Len() int {
	return len(occupancy.blocked)
}

// Occupied reports whether a position is occupied.
func (occupancy Occupancy) Occupied(position Position) bool {
	return occupancy.blocks(nodeFromPosition(position))
}

// blocks reports whether a node is occupied.
func (occupancy Occupancy) blocks(key nodeKey) bool {
	if len(occupancy.blocked) == 0 {
		return false
	}
	_, ok := occupancy.blocked[key]

	return ok
}
