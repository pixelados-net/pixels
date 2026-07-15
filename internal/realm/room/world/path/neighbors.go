package path

import "github.com/niflaot/pixels/internal/realm/room/world/grid"

// skipDirection reports whether a direction cannot be used.
func (search *search) skipDirection(current nodeKey, move direction) bool {
	if move.diagonal && search.rules.DisableDiagonal {
		return true
	}
	if !move.diagonal {
		return false
	}

	return search.diagonalBlocked(current, move)
}

// diagonalBlocked reports whether a diagonal move is blocked by both corners.
func (search *search) diagonalBlocked(current nodeKey, move direction) bool {
	horizontal, horizontalOK := neighborPoint(current, direction{dx: move.dx})
	vertical, verticalOK := neighborPoint(current, direction{dy: move.dy})
	if !horizontalOK || !verticalOK {
		return true
	}

	return !search.hasReachableSection(current, horizontal) && !search.hasReachableSection(current, vertical)
}

// hasReachableSection reports whether a neighbor has any reachable section.
func (search *search) hasReachableSection(current nodeKey, point grid.Point) bool {
	column, err := search.column(point)
	if err != nil {
		return false
	}
	for index := 0; index < column.Len(); index++ {
		section, ok := column.Section(index)
		next := nodeKey{X: section.Point().X, Y: section.Point().Y, Z: section.Z()}
		if ok && search.rules.AllowsSection(section) && column.Accepts(section) &&
			search.rules.AllowsStep(current.Z, section.Z()) &&
			!search.occupancy.blocks(next) {
			return true
		}
	}

	return false
}
