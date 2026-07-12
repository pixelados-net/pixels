// Package gate implements binary walkability-changing furniture behavior.
package gate

import (
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
)

// Behavior toggles a gate only while its complete footprint is unoccupied.
type Behavior struct{}

// Next resolves the next gate state and requests a fixture rebuild.
func (Behavior) Next(active *roomlive.Room, item worldfurniture.Item) (string, bool, bool) {
	if active.HasUnitInFurnitureFootprint(item) {
		return item.ExtraData, false, false
	}
	if item.ExtraData == "" || item.ExtraData == "0" {
		return "1", true, true
	}

	return "0", true, true
}
