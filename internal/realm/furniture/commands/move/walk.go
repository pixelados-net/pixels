package move

import (
	"context"
	"errors"

	furniturewalkedoff "github.com/niflaot/pixels/internal/realm/furniture/events/walkedoff"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/pkg/bus"
)

// stopDecoratorMovement cancels a builder's pending route before furniture manipulation.
func stopDecoratorMovement(active *roomlive.Room, playerID int64) error {
	_, err := active.StopMovement(playerID)
	if errors.Is(err, roomlive.ErrUnitNotFound) || errors.Is(err, roomlive.ErrWorldNotLoaded) {
		return nil
	}

	return err
}

// publishWalkedOff emits synthetic exits for units left behind by a moved interaction.
func (handler Handler) publishWalkedOff(ctx context.Context, active *roomlive.Room, previous worldfurniture.Item, found bool, moved worldfurniture.Item) error {
	if !found || handler.Events == nil || !previous.Definition.EmitsWalkEvents() || previous.Definition.InteractionType == "roller" {
		return nil
	}
	oldFootprint := worldfurniture.Footprint(previous.Point, previous.Definition.Width, previous.Definition.Length, previous.Rotation)
	newFootprint := worldfurniture.Footprint(moved.Point, moved.Definition.Width, moved.Definition.Length, moved.Rotation)
	var result error
	for _, unit := range active.Units() {
		if !containsPoint(oldFootprint, unit.Position.Point) || containsPoint(newFootprint, unit.Position.Point) {
			continue
		}
		result = errors.Join(result, handler.Events.Publish(ctx, bus.Event{Name: furniturewalkedoff.Name, Payload: furniturewalkedoff.Payload{
			PlayerID: unit.PlayerID, ItemID: previous.ID, RoomID: active.ID(),
		}}))
	}

	return result
}

// containsPoint reports whether one footprint contains a tile.
func containsPoint(points []grid.Point, point grid.Point) bool {
	for _, candidate := range points {
		if candidate == point {
			return true
		}
	}

	return false
}
