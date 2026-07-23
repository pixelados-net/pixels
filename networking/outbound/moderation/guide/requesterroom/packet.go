// Package requesterroom contains the moderation requesterroom outbound packet.
package requesterroom

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation requesterroom packet.
const Header uint16 = 1847

// Definition describes moderation requesterroom fields.
var Definition = codec.Definition{
	codec.Named("roomID", codec.Int32Field),
}

// Encode creates a moderation requesterroom packet.
func Encode(roomID int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(roomID))
}
