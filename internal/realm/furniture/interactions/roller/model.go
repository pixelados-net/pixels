package roller

import (
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
)

// step stores one validated roller mutation.
type step struct {
	// roller stores the source roller.
	roller worldfurniture.Item
	// target stores the adjacent destination tile.
	target grid.Point
	// offset stores the destination height delta.
	offset grid.Height
	// units stores eligible stationary player units.
	units []roomlive.UnitSnapshot
	// items stores eligible mounted furniture.
	items []worldfurniture.Item
	// sourceTop stores the source walk-hook item.
	sourceTop int64
	// targetTop stores the destination walk-hook item.
	targetTop int64
}

// movedStep stores applied mutations used by projections and persistence.
type movedStep struct {
	// step stores validated source data.
	step step
	// unitSources stores the successfully applied unit origins.
	unitSources []roomlive.UnitSnapshot
	// units stores post-mutation player snapshots.
	units []roomlive.UnitSnapshot
	// itemSources stores the successfully applied furniture origins.
	itemSources []worldfurniture.Item
	// items stores post-mutation furniture snapshots.
	items []worldfurniture.Item
}

// persistence stores one durable roller position update.
type persistence struct {
	// roomID identifies the item room.
	roomID int64
	// item stores the final world snapshot.
	item worldfurniture.Item
}
