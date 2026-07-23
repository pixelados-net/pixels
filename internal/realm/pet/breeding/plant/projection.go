package plant

import (
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	roomprojection "github.com/niflaot/pixels/internal/realm/room/runtime/projection"
	"github.com/niflaot/pixels/networking/codec"
	outadd "github.com/niflaot/pixels/networking/outbound/room/furniture/add"
)

// compostPacket maps placed RIP furniture into Nitro's incremental room packet.
func compostPacket(item furnituremodel.Item, definition furnituremodel.Definition, ownerName string) (codec.Packet, error) {
	return outadd.Encode(outadd.FloorItem{
		ID: item.ID, SpriteID: roomprojection.FurnitureSpriteID(item, definition), X: *item.X, Y: *item.Y,
		Rotation: int(item.Rotation), Z: roomprojection.FurnitureHeightValue(*item.Z), ExtraHeight: roomprojection.ExtraHeightValue(definition),
		ExtraData: item.ExtraData, Data: roomprojection.SpecializedObjectData(definition.InteractionType, item.ExtraData),
		UsagePolicy: roomprojection.UsagePolicyValue(definition), Kind: roomprojection.FurnitureKindValue(item),
		OwnerID: item.OwnerPlayerID, OwnerName: ownerName,
	})
}
