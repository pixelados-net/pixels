// Package resolution contains the moderation resolution outbound packet.
package resolution

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation resolution packet.
const Header uint16 = 2674

// Definition describes moderation resolution fields.
var Definition = codec.Definition{
	codec.Named("code", codec.Int32Field),
}

// Encode creates a moderation resolution packet.
func Encode(code int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(code))
}
