package floorplan

import (
	"context"
	"errors"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	floorplansaved "github.com/niflaot/pixels/internal/realm/room/control/events/floorplansaved"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	roomlayout "github.com/niflaot/pixels/internal/realm/room/world/layout"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	outrefresh "github.com/niflaot/pixels/networking/outbound/inventory/furniture/refresh"
	outunseen "github.com/niflaot/pixels/networking/outbound/inventory/unseen"
	outforward "github.com/niflaot/pixels/networking/outbound/room/forward"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/zap"
)

// reload replaces active geometry and asks Nitro to rebuild its room renderer.
func (handler SaveHandler) reload(ctx context.Context, room roommodel.Room, active *roomlive.Room, roomLayout roomlayout.Layout) error {
	roomGrid, err := roomLayout.Grid()
	if err != nil {
		return err
	}
	door, ok := grid.NewPoint(roomLayout.DoorX, roomLayout.DoorY)
	if !ok {
		return roomlive.ErrInvalidWorld
	}
	furniture, err := handler.loadWorldFurniture(ctx, room.ID)
	if err != nil {
		return err
	}
	rotation := worldunit.Rotation(roomLayout.DoorDirection % 8)
	if err = active.LoadWorld(roomlive.WorldConfig{
		Grid: roomGrid, Furniture: furniture,
		Door: worldpath.Position{Point: door, Z: grid.HeightFromInt(roomLayout.DoorZ)},
		Body: rotation, Head: rotation, Rules: worldpath.DefaultRules(),
	}); err != nil {
		return err
	}
	packet, err := outforward.Encode(int32(room.ID))
	if err != nil {
		return err
	}

	return broadcast.RoomPacket(ctx, handler.Connections, active, packet, 0)
}

// notifyPicked refreshes inventories for online owners after commit.
func (handler SaveHandler) notifyPicked(ctx context.Context, picked []furnituremodel.Item) {
	byOwner := make(map[int64][]int64)
	for _, item := range picked {
		byOwner[item.OwnerPlayerID] = append(byOwner[item.OwnerPlayerID], item.ID)
	}
	for ownerID, itemIDs := range byOwner {
		bindingValue, found := handler.Bindings.FindByPlayer(ownerID)
		if !found {
			continue
		}
		connection, found := handler.Connections.Get(bindingValue.ConnectionKind, bindingValue.ConnectionID)
		if !found {
			continue
		}
		unseen, err := outunseen.EncodeOwned(itemIDs)
		if err == nil {
			err = connection.Send(ctx, unseen)
		}
		refresh, encodeErr := outrefresh.Encode()
		if encodeErr == nil {
			err = errors.Join(err, connection.Send(ctx, refresh))
		}
		if err != nil && handler.Log != nil {
			handler.Log.Warn("floor plan inventory refresh failed", zap.Int64("player_id", ownerID), zap.Error(err))
		}
	}
}

// publish emits a committed floor plan event.
func (handler SaveHandler) publish(ctx context.Context, roomID int64, actorID int64) error {
	if handler.Events == nil {
		return nil
	}

	return handler.Events.Publish(ctx, bus.Event{Name: floorplansaved.Name, Payload: floorplansaved.Payload{RoomID: roomID, ActorID: actorID}})
}
