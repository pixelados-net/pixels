// Package remove contains the REMOVE_FURNITURE_FROM_INVENTORY outbound packet.
package remove

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the REMOVE_FURNITURE_FROM_INVENTORY packet identifier.
	Header uint16 = 159
)

// Definition describes the REMOVE_FURNITURE_FROM_INVENTORY payload fields.
var Definition = codec.Definition{codec.Named("id", codec.Int32Field)}

// Encode creates a REMOVE_FURNITURE_FROM_INVENTORY packet.
func Encode(itemID int64) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(int32(itemID)))
}
