// Package state contains the FURNITURE_STATE outbound packet.
package state

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the FURNITURE_STATE packet identifier.
	Header uint16 = 2376
)

// Definition describes the FURNITURE_STATE payload fields.
var Definition = codec.Definition{
	codec.Named("itemId", codec.Int32Field),
	codec.Named("state", codec.Int32Field),
}

// Encode creates a FURNITURE_STATE packet.
func Encode(itemID int64, value int) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(int32(itemID)), codec.Int32(int32(value)))
}
