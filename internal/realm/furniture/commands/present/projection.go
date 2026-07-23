package present

import (
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	"github.com/niflaot/pixels/internal/realm/room/runtime/projection"
	"github.com/niflaot/pixels/networking/codec"
	outadd "github.com/niflaot/pixels/networking/outbound/room/furniture/add"
)

// addPacket maps opened furniture into an ADD_FLOOR_ITEM packet.
func addPacket(item furnituremodel.Item, definition furnituremodel.Definition, ownerName string) (codec.Packet, error) {
	return outadd.Encode(outadd.FloorItem{
		ID: item.ID, SpriteID: projection.FurnitureSpriteID(item, definition),
		X: *item.X, Y: *item.Y, Rotation: int(item.Rotation),
		Z: projection.FurnitureHeightValue(*item.Z), ExtraHeight: projection.ExtraHeightValue(definition),
		ExtraData: item.ExtraData, UsagePolicy: projection.UsagePolicyValue(definition),
		Kind: projection.FurnitureKindValue(item), GiftWrapped: item.GiftWrapped,
		OwnerID: item.OwnerPlayerID, OwnerName: ownerName,
	})
}
