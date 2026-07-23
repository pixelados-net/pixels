package unit

import (
	"time"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
)

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

// SetFloorPosture changes the free-standing sit posture and stops dancing.
func (unit *Unit) SetFloorPosture(sitting bool) {
	unit.statuses.clear(StatusLay)
	unit.statuses.clear(StatusDance)
	if sitting {
		unit.statuses.set(StatusSit, "0.0")
		return
	}
	unit.statuses.clear(StatusSit)
}

// SetIdle replaces the unit AFK projection.
func (unit *Unit) SetIdle(idle bool) {
	unit.SetIdleAt(idle, time.Now())
}

// SetIdleAt replaces the unit AFK projection at one deterministic instant.
func (unit *Unit) SetIdleAt(idle bool, at time.Time) {
	unit.setIdleAt(idle, false, at)
}

// SetManualIdleAt replaces the unit manual AFK projection at one deterministic instant.
func (unit *Unit) SetManualIdleAt(idle bool, at time.Time) {
	unit.setIdleAt(idle, idle, at)
}

// setIdleAt replaces the unit AFK projection and its source.
func (unit *Unit) setIdleAt(idle bool, manual bool, at time.Time) {
	unit.idle = idle
	unit.idleSince = time.Time{}
	unit.manualIdle = false
	if idle {
		unit.idleSince = at
		unit.manualIdle = manual
	}
}

// Idle reports whether the unit is projected as AFK.
func (unit *Unit) Idle() bool {
	return unit.idle
}

// IdleSince returns when the current AFK projection began.
func (unit *Unit) IdleSince() time.Time {
	return unit.idleSince
}

// ManualIdle reports whether explicit avatar activity must clear the AFK projection.
func (unit *Unit) ManualIdle() bool {
	return unit.manualIdle
}

// SetActiveEffect replaces the selected avatar effect.
func (unit *Unit) SetActiveEffect(effectID int32) {
	unit.activeEffectID = effectID
}

// ActiveEffect returns the selected avatar effect.
func (unit *Unit) ActiveEffect() int32 {
	return unit.activeEffectID
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
