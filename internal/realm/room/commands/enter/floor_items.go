package enter

import (
	"context"

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

	owners, records := projection.FloorItems(items, definitions, ownerNames(room, active))
	packet, err := outflooritems.Encode(owners, records)
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}

// ownerNames resolves best-effort owner display names from the room owner and current occupants.
func ownerNames(room roommodel.Room, active *roomlive.Room) map[int64]string {
	names := map[int64]string{room.OwnerPlayerID: room.OwnerName}
	if active == nil {
		return names
	}
	for _, occupant := range active.Occupants() {
		names[occupant.PlayerID] = occupant.Username
	}

	return names
}
