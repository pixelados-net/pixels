// Package dicevalue contains the FURNITURE_STATE_2 outbound packet.
package dicevalue

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the FURNITURE_STATE_2 packet identifier.
	Header uint16 = 3431
)

// Definition describes the FURNITURE_STATE_2 payload fields.
var Definition = codec.Definition{
	codec.Named("itemId", codec.Int32Field),
	codec.Named("value", codec.Int32Field),
}

// Encode creates a FURNITURE_STATE_2 packet.
func Encode(itemID int64, value int) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(int32(itemID)), codec.Int32(int32(value)))
}
