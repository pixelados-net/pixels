// Package sanction contains the moderation sanction outbound packet.
package sanction

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation sanction packet.
const Header uint16 = 2782

// Definition describes moderation sanction fields.
var Definition = codec.Definition{
	codec.Named("kind", codec.Int32Field),
	codec.Named("hours", codec.Int32Field),
}

// Encode creates a moderation sanction packet.
func Encode(kind int32, hours int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(kind), codec.Int32(hours))
}
