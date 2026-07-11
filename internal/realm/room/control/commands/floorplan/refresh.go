package floorplan

import (
	"context"
	"errors"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	floorplansaved "github.com/niflaot/pixels/internal/realm/room/control/events/floorplansaved"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/runtime/projection"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	roomitems "github.com/niflaot/pixels/internal/realm/room/world/items"
	roomlayout "github.com/niflaot/pixels/internal/realm/room/world/layout"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/networking/codec"
	outrefresh "github.com/niflaot/pixels/networking/outbound/inventory/furniture/refresh"
	outunseen "github.com/niflaot/pixels/networking/outbound/inventory/unseen"
	outstatus "github.com/niflaot/pixels/networking/outbound/room/entities/status"
	outunits "github.com/niflaot/pixels/networking/outbound/room/entities/units"
	outentrytile "github.com/niflaot/pixels/networking/outbound/room/entrytile"
	outflooritems "github.com/niflaot/pixels/networking/outbound/room/furniture/flooritems"
	outheightmap "github.com/niflaot/pixels/networking/outbound/room/heightmap"
	outmodel "github.com/niflaot/pixels/networking/outbound/room/model"
	outthickness "github.com/niflaot/pixels/networking/outbound/room/thickness/updated"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/zap"
)

// reload replaces active geometry and broadcasts a complete room-world refresh.
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
		Door: worldpath.Position{Point: door, Z: grid.Height(roomLayout.DoorZ)},
		Body: rotation, Head: rotation, Rules: worldpath.DefaultRules(),
	}); err != nil {
		return err
	}
	packets, err := handler.refreshPackets(ctx, room, active, roomLayout)
	if err != nil {
		return err
	}
	for _, packet := range packets {
		_ = broadcast.RoomPacket(ctx, handler.Connections, active, packet, 0)
	}

	return nil
}

// refreshPackets encodes a complete active room-world projection in client order.
func (handler SaveHandler) refreshPackets(ctx context.Context, room roommodel.Room, active *roomlive.Room, roomLayout roomlayout.Layout) ([]codec.Packet, error) {
	entry, err := outentrytile.Encode(int32(roomLayout.DoorX), int32(roomLayout.DoorY), formatHeight(roomLayout.DoorZ), int32(roomLayout.DoorDirection))
	if err != nil {
		return nil, err
	}
	model, err := outmodel.Encode(true, int32(roomLayout.WallHeight), roomLayout.Heightmap)
	if err != nil {
		return nil, err
	}
	wall, floor := visualizationThickness(room, roomLayout)
	thickness, err := outthickness.Encode(room.HideWalls, int32(wall), int32(floor))
	if err != nil {
		return nil, err
	}
	floorItems, err := handler.floorItemsPacket(ctx, room, active)
	if err != nil {
		return nil, err
	}
	width, _, tiles := active.SurfaceHeights()
	heights, err := outheightmap.Encode(int(width), projection.HeightMapTiles(tiles))
	if err != nil {
		return nil, err
	}
	units, err := outunits.Encode(projection.Units(active))
	if err != nil {
		return nil, err
	}
	statuses, err := outstatus.Encode(projection.Statuses(active))
	if err != nil {
		return nil, err
	}

	return []codec.Packet{entry, model, thickness, floorItems, heights, units, statuses}, nil
}

// floorItemsPacket encodes every persistent floor item after optional pickups.
func (handler SaveHandler) floorItemsPacket(ctx context.Context, room roommodel.Room, active *roomlive.Room) (codec.Packet, error) {
	items, err := handler.Furniture.ListRoomItems(ctx, room.ID)
	if err != nil {
		return codec.Packet{}, err
	}
	definitions, err := roomitems.DefinitionsByID(ctx, handler.Furniture)
	if err != nil {
		return codec.Packet{}, err
	}
	names := map[int64]string{room.OwnerPlayerID: room.OwnerName}
	for _, occupant := range active.Occupants() {
		names[occupant.PlayerID] = occupant.Username
	}
	owners, records := projection.FloorItems(items, definitions, names)

	return outflooritems.Encode(owners, records)
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
