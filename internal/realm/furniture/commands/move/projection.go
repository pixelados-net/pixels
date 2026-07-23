package move

import (
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	"github.com/niflaot/pixels/internal/realm/room/runtime/projection"
	outupdate "github.com/niflaot/pixels/networking/outbound/room/furniture/update"
)

// updateRecord maps a moved item and its definition into a FLOOR_ITEM_UPDATE record.
func updateRecord(item furnituremodel.Item, definition furnituremodel.Definition) outupdate.FloorItem {
	return outupdate.FloorItem{
		ID: item.ID, SpriteID: definition.SpriteID, X: *item.X, Y: *item.Y,
		Rotation: int(item.Rotation), Z: projection.FurnitureHeightValue(*item.Z),
		ExtraHeight: projection.ExtraHeightValue(definition), ExtraData: item.ExtraData,
		OwnerID: item.OwnerPlayerID,
	}
}
