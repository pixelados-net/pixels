// Package offered contains the moderation offered outbound packet.
package offered

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation offered packet.
const Header uint16 = 735

// Definition describes moderation offered fields.
var Definition = codec.Definition{
	codec.Named("ticketID", codec.Int32Field),
}

// Encode creates a moderation offered packet.
func Encode(ticketID int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(ticketID))
}
