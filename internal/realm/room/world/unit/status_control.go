package unit

import "github.com/niflaot/pixels/internal/realm/room/world/grid"

// SetStatus stores a unit status value.
func (unit *Unit) SetStatus(key string, value string) {
	unit.statuses.set(key, value)
}

// ClearStatus removes a unit status value.
func (unit *Unit) ClearStatus(key string) {
	unit.statuses.clear(key)
}

// HasStatus reports whether the unit currently has a status key.
func (unit *Unit) HasStatus(key string) bool {
	return unit.statuses.has(key)
}

// Settled reports whether the unit currently sits or lays on a slot.
func (unit *Unit) Settled() bool {
	return unit.statuses.has(StatusSit) || unit.statuses.has(StatusLay)
}

// StandUp clears sit and lay statuses, returning a settled unit to standing in place.
func (unit *Unit) StandUp() {
	unit.statuses.clear(StatusSit)
	unit.statuses.clear(StatusLay)
}

// SetHeight corrects the unit's vertical position without moving it off its current tile, used when
// the surface underneath changes height (e.g. the furniture a unit stood on moved away) so a later
// path search validates against where the unit now actually stands instead of a stale section.
func (unit *Unit) SetHeight(z grid.Height) {
	unit.position.Z = z
}

// Statuses returns current statuses ordered by key.
func (unit *Unit) Statuses() []Status {
	return unit.statuses.snapshot()
}

// SetHandItem replaces the currently carried hand item.
func (unit *Unit) SetHandItem(itemID int32) {
	unit.handItem = itemID
}

// HandItem returns the currently carried hand item.
func (unit *Unit) HandItem() int32 {
	return unit.handItem
}
