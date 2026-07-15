// Package ended contains the moderation ended outbound packet.
package ended

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation ended packet.
const Header uint16 = 1456

// Definition describes moderation ended fields.
var Definition = codec.Definition{
	codec.Named("reason", codec.Int32Field),
}

// Encode creates a moderation ended packet.
func Encode(reason int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(reason))
}
