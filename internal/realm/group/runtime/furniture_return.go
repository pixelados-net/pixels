package runtime

import (
	"context"

	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	roombroadcast "github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/networking/codec"
	outrefresh "github.com/niflaot/pixels/networking/outbound/inventory/furniture/refresh"
	outunseen "github.com/niflaot/pixels/networking/outbound/inventory/unseen"
	outremove "github.com/niflaot/pixels/networking/outbound/room/furniture/remove"
	outwallremove "github.com/niflaot/pixels/networking/outbound/room/furniture/wallremove"
)

// ReturnFurniture removes committed headquarters returns from the active room and refreshes the
// receiving player's inventory.
func (projector *Projector) ReturnFurniture(ctx context.Context, returned grouprecord.FurnitureReturn) error {
	if projector == nil || len(returned.Items) == 0 {
		return nil
	}
	roomErr := projector.removeReturnedFurniture(ctx, returned)
	inventoryErr := projector.refreshReturnedInventory(ctx, returned.Items)
	if roomErr != nil {
		return roomErr
	}
	return inventoryErr
}

// removeReturnedFurniture reconciles the active room world and every occupant's visual state.
func (projector *Projector) removeReturnedFurniture(ctx context.Context, returned grouprecord.FurnitureReturn) error {
	if projector.rooms == nil {
		return nil
	}
	active, found := projector.rooms.Find(returned.RoomID)
	if !found {
		return nil
	}
	points := make([]grid.Point, 0, len(returned.Items))
	stoodUp := make([]roomlive.UnitSnapshot, 0)
	for _, item := range returned.Items {
		if !item.Wall {
			worldItem, loaded := active.FurnitureItem(item.ItemID)
			if loaded {
				units, err := active.ReloadFurniture(item.ItemID, nil)
				if err != nil {
					return err
				}
				stoodUp = append(stoodUp, units...)
				footprint := worldfurniture.Footprint(
					worldItem.Point,
					worldItem.Definition.Width,
					worldItem.Definition.Length,
					worldItem.Rotation,
				)
				points = append(points, footprint...)
			}
		}
		packet, err := returnedFurnitureRemovePacket(item)
		if err != nil {
			return err
		}
		if err = roombroadcast.RoomPacket(ctx, projector.connections, active, packet, 0); err != nil {
			return err
		}
	}
	if err := roombroadcast.RoomUnitStatuses(ctx, projector.connections, active, stoodUp, 0); err != nil {
		return err
	}
	return roombroadcast.RoomHeightMapUpdate(ctx, projector.connections, active, points, 0)
}

// refreshReturnedInventory invalidates the online owner's furniture inventory after commit.
func (projector *Projector) refreshReturnedInventory(ctx context.Context, items []grouprecord.ReturnedFurniture) error {
	if projector.delivery == nil || len(items) == 0 {
		return nil
	}
	itemIDs := make([]int64, len(items))
	for index, item := range items {
		itemIDs[index] = item.ItemID
	}
	packet, err := outunseen.EncodeOwned(itemIDs)
	if err != nil {
		return err
	}
	if _, err = projector.delivery.Send(ctx, items[0].OwnerPlayerID, packet); err != nil {
		return err
	}
	packet, err = outrefresh.Encode()
	if err != nil {
		return err
	}
	_, err = projector.delivery.Send(ctx, items[0].OwnerPlayerID, packet)
	return err
}

// returnedFurnitureRemovePacket encodes the protocol variant for one returned item.
func returnedFurnitureRemovePacket(item grouprecord.ReturnedFurniture) (codec.Packet, error) {
	if item.Wall {
		return outwallremove.Encode(item.ItemID, item.OwnerPlayerID)
	}
	return outremove.Encode(item.ItemID, item.OwnerPlayerID)
}
