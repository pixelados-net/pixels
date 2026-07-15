// Package postitplaced confirms post-it placement to inventory.
package postitplaced

import "github.com/niflaot/pixels/networking/codec"

// Header is the USER_FURNITURE_POSTIT_PLACED identifier.
const Header uint16 = 1501

// Encode creates a USER_FURNITURE_POSTIT_PLACED packet.
func Encode(itemID int64, itemsLeft int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(int32(itemID)), codec.Int32(itemsLeft))
}
