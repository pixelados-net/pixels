// Package handitem contains the UNIT_HAND_ITEM outbound packet.
package handitem

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the UNIT_HAND_ITEM packet identifier.
	Header uint16 = 1474
)

// Definition describes the UNIT_HAND_ITEM payload fields.
var Definition = codec.Definition{
	codec.Named("roomIndex", codec.Int32Field),
	codec.Named("itemId", codec.Int32Field),
}

// Encode creates a UNIT_HAND_ITEM packet.
func Encode(roomIndex int64, itemID int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(int32(roomIndex)), codec.Int32(itemID))
}
