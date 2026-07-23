// Package opened contains the PRESENT_OPENED outbound packet.
package opened

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the PRESENT_OPENED packet identifier.
	Header uint16 = 56

	// FloorItemType identifies floor furniture in Nitro product packets.
	FloorItemType = "s"
)

// Definition describes the PRESENT_OPENED payload fields.
var Definition = codec.Definition{
	codec.Named("itemType", codec.StringField),
	codec.Named("classId", codec.Int32Field),
	codec.Named("productCode", codec.StringField),
	codec.Named("placedItemId", codec.Int32Field),
	codec.Named("placedItemType", codec.StringField),
	codec.Named("placedInRoom", codec.BooleanField),
	codec.Named("petFigureString", codec.StringField),
}

// Encode creates a PRESENT_OPENED packet for floor furniture.
func Encode(classID int32, productCode string, placedItemID int64, placedInRoom bool) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition,
		codec.String(FloorItemType),
		codec.Int32(classID),
		codec.String(productCode),
		codec.Int32(int32(placedItemID)),
		codec.String(FloorItemType),
		codec.Bool(placedInRoom),
		codec.String(""),
	)
}
