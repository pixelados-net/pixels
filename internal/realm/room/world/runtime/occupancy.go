package runtime

import "github.com/niflaot/pixels/internal/realm/room/world/grid"

// HasUnitAt reports whether any player, bot, or pet currently owns a tile.
func (world *World) HasUnitAt(point grid.Point) bool {
	for _, roomUnit := range world.units {
		if roomUnit.Position().Point == point {
			return true
		}
	}
	return false
}
