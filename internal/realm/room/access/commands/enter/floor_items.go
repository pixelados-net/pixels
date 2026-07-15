package enter

import (
	"context"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/runtime/projection"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	roomfurniture "github.com/niflaot/pixels/internal/realm/room/world/items"
	"github.com/niflaot/pixels/internal/realm/room/world/layout"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	netconn "github.com/niflaot/pixels/networking/connection"
	outflooritems "github.com/niflaot/pixels/networking/outbound/room/furniture/flooritems"
	outwallitems "github.com/niflaot/pixels/networking/outbound/room/furniture/wallitems"
	outpaint "github.com/niflaot/pixels/networking/outbound/room/paint"
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
	senders, err := handler.giftSenders(ctx, active, items)
	if err != nil {
		return err
	}

	owners, records := projection.FloorItems(items, definitions, names, senders)
	packet, err := outflooritems.Encode(owners, records)
	if err != nil {
		return err
	}
	if len(records) > 0 {
		if err := connection.Send(ctx, packet); err != nil {
			return err
		}
	}

	wallOwners, wallRecords := projection.WallItems(items, definitions, names)
	if len(wallRecords) == 0 {
		return nil
	}
	packet, err = outwallitems.Encode(wallOwners, wallRecords)
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}

// sendAppearance projects durable room surfaces before unit rendering.
func sendAppearance(ctx context.Context, connection netconn.Context, room roommodel.Room) error {
	for _, surface := range [][2]string{{"floor", room.FloorPaint}, {"wallpaper", room.Wallpaper}, {"landscape", room.Landscape}} {
		if surface[1] == "" {
			continue
		}
		packet, err := outpaint.Encode(surface[0], surface[1])
		if err != nil {
			return err
		}
		if err = connection.Send(ctx, packet); err != nil {
			return err
		}
	}

	return nil
}

// giftSenders resolves visible sender data for placed gifts.
func (handler Handler) giftSenders(ctx context.Context, active *roomlive.Room, items []furnituremodel.Item) (map[int64]projection.GiftSender, error) {
	senders := make(map[int64]projection.GiftSender)
	if active != nil {
		for _, occupant := range active.Occupants() {
			senders[occupant.PlayerID] = projection.GiftSender{Name: occupant.Username, Figure: occupant.Figure}
		}
	}
	if handler.PlayerDirectory == nil {
		return senders, nil
	}

	for _, item := range items {
		if item.GiftSenderPlayerID == nil {
			continue
		}
		if _, resolved := senders[*item.GiftSenderPlayerID]; resolved {
			continue
		}
		record, found, err := handler.PlayerDirectory.FindByID(ctx, *item.GiftSenderPlayerID)
		if err != nil {
			return nil, err
		}
		if found {
			senders[*item.GiftSenderPlayerID] = projection.GiftSender{
				Name: record.Player.Username, Figure: record.Profile.Look,
			}
		}
	}

	return senders, nil
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

// loadWorld loads the runtime world from persistent layout and furniture state.
func (handler Handler) loadWorld(ctx context.Context, room *roomlive.Room, roomData roommodel.Room, roomLayout layout.Layout) error {
	roomGrid, err := roomLayout.Grid()
	if err != nil {
		return err
	}
	doorPoint, ok := grid.NewPoint(roomLayout.DoorX, roomLayout.DoorY)
	if !ok {
		return roomlive.ErrInvalidWorld
	}
	furnitureItems, err := handler.loadFurniture(ctx, roomData.ID)
	if err != nil {
		return err
	}
	rotation := worldunit.Rotation(roomLayout.DoorDirection % 8)

	return room.LoadWorld(roomlive.WorldConfig{
		Grid: roomGrid, Furniture: furnitureItems,
		Door: worldpath.Position{Point: doorPoint, Z: grid.HeightFromInt(roomLayout.DoorZ)},
		Body: rotation, Head: rotation, Rules: worldpath.DefaultRules(),
	})
}

// loadFurniture loads placed furniture when a manager is configured.
func (handler Handler) loadFurniture(ctx context.Context, roomID int64) ([]worldfurniture.Item, error) {
	if handler.Furniture == nil {
		return nil, nil
	}

	return roomfurniture.LoadRoomFurniture(ctx, handler.Furniture, roomID)
}
