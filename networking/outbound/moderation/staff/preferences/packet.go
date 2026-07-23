// Package preferences contains the moderation preferences outbound packet.
package preferences

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation preferences packet.
const Header uint16 = 1576

// Definition describes moderation preferences fields.
var Definition = codec.Definition{
	codec.Named("x", codec.Int32Field),
	codec.Named("y", codec.Int32Field),
	codec.Named("width", codec.Int32Field),
	codec.Named("height", codec.Int32Field),
}

// Encode creates a moderation preferences packet.
func Encode(x int32, y int32, width int32, height int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(x), codec.Int32(y), codec.Int32(width), codec.Int32(height))
}
