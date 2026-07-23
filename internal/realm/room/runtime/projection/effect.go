package projection

import (
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// EffectID returns the visible room effect without discarding the selected effect underneath riding.
func EffectID(unit roomlive.UnitSnapshot) int32 {
	if unit.RenderOffset == worldunit.RidingHeightOffset {
		return worldunit.RidingEffectID
	}

	return unit.ActiveEffectID
}
