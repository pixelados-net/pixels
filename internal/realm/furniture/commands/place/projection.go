package place

import (
	"context"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	"github.com/niflaot/pixels/internal/realm/room/runtime/projection"
	outadd "github.com/niflaot/pixels/networking/outbound/room/furniture/add"
)

// giftSender resolves the visible sender for wrapped gift tags.
func (handler Handler) giftSender(ctx context.Context, item furnituremodel.Item) (projection.GiftSender, error) {
	if item.GiftSenderPlayerID == nil {
		return projection.GiftSender{}, nil
	}
	if handler.Players != nil {
		if player, found := handler.Players.Find(*item.GiftSenderPlayerID); found {
			snapshot := player.Snapshot()
			return projection.GiftSender{Name: snapshot.Username, Figure: snapshot.Look}, nil
		}
	}
	if handler.PlayerDirectory == nil {
		return projection.GiftSender{}, nil
	}
	record, found, err := handler.PlayerDirectory.FindByID(ctx, *item.GiftSenderPlayerID)
	if err != nil || !found {
		return projection.GiftSender{}, err
	}

	return projection.GiftSender{Name: record.Player.Username, Figure: record.Profile.Look}, nil
}

// addRecord maps a placed item and its definition into an ADD_FLOOR_ITEM record.
func addRecord(item furnituremodel.Item, definition furnituremodel.Definition, ownerName string, sender projection.GiftSender) outadd.FloorItem {
	record := outadd.FloorItem{
		ID: item.ID, SpriteID: projection.FurnitureSpriteID(item, definition),
		X: *item.X, Y: *item.Y, Rotation: int(item.Rotation),
		Z: projection.FurnitureHeightValue(*item.Z), ExtraHeight: projection.ExtraHeightValue(definition),
		ExtraData: item.ExtraData, UsagePolicy: projection.UsagePolicyValue(definition),
		Kind: projection.FurnitureKindValue(item), GiftWrapped: item.GiftWrapped,
		OwnerID: item.OwnerPlayerID, OwnerName: ownerName, GiftMessage: giftMessage(item),
		GiftProductCode: definition.Name, GiftSenderName: sender.Name, GiftSenderFigure: sender.Figure,
	}
	record.Data = projection.SpecializedObjectData(definition.InteractionType, item.ExtraData)

	return record
}

// giftMessage returns the optional persisted gift message.
func giftMessage(item furnituremodel.Item) string {
	if item.GiftMessage == nil {
		return ""
	}

	return *item.GiftMessage
}
