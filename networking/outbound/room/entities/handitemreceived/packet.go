// Package handitemreceived encodes the private hand-item receipt notice.
package handitemreceived

import "github.com/niflaot/pixels/networking/codec"

// Header is the HAND_ITEM_RECEIVED identifier.
const Header uint16 = 354

// Encode creates a HAND_ITEM_RECEIVED packet.
func Encode(giverUnitID int64, itemID int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(int32(giverUnitID)), codec.Int32(itemID))
}
