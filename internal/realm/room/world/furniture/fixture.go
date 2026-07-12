package furniture

import (
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/internal/realm/room/world/surface"
)

// Fixtures converts a placed item into resolver fixtures, one per occupied tile.
func Fixtures(item Item) ([]surface.Fixture, error) {
	footprint := Footprint(item.Point, item.Definition.Width, item.Definition.Length, item.Rotation)
	slotsByPoint := indexSlotsByPoint(Slots(item))
	top := item.Top()

	fixtures := make([]surface.Fixture, 0, len(footprint))
	for _, point := range footprint {
		fixture, err := tileFixture(item, point, top, slotsByPoint)
		if err != nil {
			return nil, err
		}
		fixtures = append(fixtures, fixture)
	}

	return fixtures, nil
}

// tileFixture builds the resolver fixture for one footprint tile. Slot tiles expose their walkable
// section at the item's base height, not its top: units reach a seat or bed at floor level and the
// sit/lay status carries the visual height offset, so pathfinding step limits never gate on how tall
// the furniture is.
func tileFixture(item Item, point grid.Point, top grid.Height, slotsByPoint map[grid.Point]Slot) (surface.Fixture, error) {
	if slot, ok := slotsByPoint[point]; ok {
		return surface.NewFixture(surface.FixtureParams{
			Point:    point,
			Z:        item.Z,
			Top:      top,
			Stacking: item.Definition.AllowStack,
			State:    slotState(slot.Status),
			SourceID: item.ID,
		})
	}

	if gateOpen(item) {
		return surface.NewFixture(surface.FixtureParams{
			Point:    point,
			Z:        item.Z,
			Top:      top,
			Stacking: item.Definition.AllowStack,
			State:    surface.StateOpen,
			SourceID: item.ID,
		})
	}

	if item.Definition.AllowWalk {
		return surface.NewFixture(surface.FixtureParams{
			Point:    point,
			Z:        top,
			Top:      top,
			Stacking: item.Definition.AllowStack,
			State:    surface.StateOpen,
			SourceID: item.ID,
		})
	}

	return surface.NewFixture(surface.FixtureParams{
		Point:    point,
		Z:        item.Z,
		Top:      top,
		Stacking: item.Definition.AllowStack,
		State:    surface.StateBlocked,
		SourceID: item.ID,
	})
}

// gateOpen reports whether a gate exposes a walkable fixture.
func gateOpen(item Item) bool {
	return item.Definition.InteractionType == "gate" && item.ExtraData == "1"
}

// slotState maps a slot status to its resolver section state.
func slotState(status SlotStatus) surface.State {
	if status == SlotStatusLay {
		return surface.StateLay
	}

	return surface.StateSit
}

// indexSlotsByPoint indexes resolved slots by their absolute tile.
func indexSlotsByPoint(slots []Slot) map[grid.Point]Slot {
	indexed := make(map[grid.Point]Slot, len(slots))
	for _, slot := range slots {
		indexed[slot.Point] = slot
	}

	return indexed
}
