// Package creationresult contains the moderation creationresult outbound packet.
package creationresult

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation creationresult packet.
const Header uint16 = 3285

// Definition describes moderation creationresult fields.
var Definition = codec.Definition{
	codec.Named("code", codec.Int32Field),
}

// Encode creates a moderation creationresult packet.
func Encode(code int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(code))
}
