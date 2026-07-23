package projection

import (
	"testing"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// TestEffectIDPreservesRidingProjection verifies selected effects remain hidden until dismount.
func TestEffectIDPreservesRidingProjection(t *testing.T) {
	unit := roomlive.UnitSnapshot{ActiveEffectID: 8, RenderOffset: worldunit.RidingHeightOffset}
	if effectID := EffectID(unit); effectID != worldunit.RidingEffectID {
		t.Fatalf("effect=%d", effectID)
	}
	unit.RenderOffset = 0
	if effectID := EffectID(unit); effectID != 8 {
		t.Fatalf("restored effect=%d", effectID)
	}
}
