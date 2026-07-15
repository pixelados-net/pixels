// Package typing contains the moderation typing outbound packet.
package typing

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation typing packet.
const Header uint16 = 1016

// Definition describes moderation typing fields.
var Definition = codec.Definition{
	codec.Named("typing", codec.BooleanField),
}

// Encode creates a moderation typing packet.
func Encode(typing bool) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Bool(typing))
}
