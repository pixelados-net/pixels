// Package add contains the ADD_FURNITURE_TO_INVENTORY outbound packet.
package add

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the ADD_FURNITURE_TO_INVENTORY packet identifier.
	Header uint16 = 2103

	// ownedFurniCategory is the inventory category number for owned furniture.
	ownedFurniCategory int32 = 1
)

// Definition describes the ADD_FURNITURE_TO_INVENTORY payload fields.
var Definition = codec.Definition{
	codec.Named("categoryCount", codec.Int32Field),
	codec.Named("category", codec.Int32Field),
	codec.Named("idCount", codec.Int32Field),
	codec.Named("id", codec.Int32Field),
}

// Encode creates an ADD_FURNITURE_TO_INVENTORY packet for one item returning to inventory.
func Encode(itemID int64) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition,
		codec.Int32(1),
		codec.Int32(ownedFurniCategory),
		codec.Int32(1),
		codec.Int32(int32(itemID)),
	)
}
