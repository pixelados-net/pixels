// Package error contains the moderation error outbound packet.
package error

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation error packet.
const Header uint16 = 673

// Definition describes moderation error fields.
var Definition = codec.Definition{
	codec.Named("code", codec.Int32Field),
}

// Encode creates a moderation error packet.
func Encode(code int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(code))
}
