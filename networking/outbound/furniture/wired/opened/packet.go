// Package opened encodes the WIRED_OPEN outbound packet.
package opened

import "github.com/niflaot/pixels/networking/codec"

// Header is the WIRED_OPEN packet identifier.
const Header uint16 = 1830

// Definition describes WIRED_OPEN.
var Definition = codec.Definition{codec.Named("itemId", codec.Int32Field)}

// Encode creates a WIRED opened packet.
func Encode(itemID int64) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(int32(itemID)))
}
