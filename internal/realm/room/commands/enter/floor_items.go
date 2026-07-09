package enter

import (
	"context"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	roomfurniture "github.com/niflaot/pixels/internal/realm/room/furniture"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
	"github.com/niflaot/pixels/internal/realm/room/projection"
	netconn "github.com/niflaot/pixels/networking/connection"
	outflooritems "github.com/niflaot/pixels/networking/outbound/room/furniture/flooritems"
)

// sendFloorItems sends the current room floor items snapshot to one connection.
func (handler Handler) sendFloorItems(ctx context.Context, connection netconn.Context, room roommodel.Room, active *roomlive.Room) error {
	if handler.Furniture == nil {
		return nil
	}

	items, err := handler.Furniture.ListRoomItems(ctx, room.ID)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}

	definitions, err := roomfurniture.DefinitionsByID(ctx, handler.Furniture)
	if err != nil {
		return err
	}

	names, err := handler.ownerNames(ctx, room, active, items)
	if err != nil {
		return err
	}

	owners, records := projection.FloorItems(items, definitions, names)
	packet, err := outflooritems.Encode(owners, records)
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}

// ownerNames resolves owner display names from the room owner, current occupants, and, for any
// remaining item owner not currently online, the durable player directory. Without a resolvable
// name the client falls back to displaying the literal string "null", so every owner id present
// among the room's items must end up with an entry here.
func (handler Handler) ownerNames(ctx context.Context, room roommodel.Room, active *roomlive.Room, items []furnituremodel.Item) (map[int64]string, error) {
	names := map[int64]string{room.OwnerPlayerID: room.OwnerName}
	if active != nil {
		for _, occupant := range active.Occupants() {
			names[occupant.PlayerID] = occupant.Username
		}
	}

	if handler.PlayerDirectory == nil {
		return names, nil
	}

	for _, item := range items {
		if _, resolved := names[item.OwnerPlayerID]; resolved {
			continue
		}

		record, found, err := handler.PlayerDirectory.FindByID(ctx, item.OwnerPlayerID)
		if err != nil {
			return nil, err
		}
		if found {
			names[item.OwnerPlayerID] = record.Player.Username
		}
	}

	return names, nil
}
