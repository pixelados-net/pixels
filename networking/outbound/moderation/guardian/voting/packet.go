// Package voting contains the moderation voting outbound packet.
package voting

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation voting packet.
const Header uint16 = 1829

// Definition describes moderation voting fields.
var Definition = codec.Definition{
	codec.Named("ticketID", codec.Int32Field),
	codec.Named("secondsLeft", codec.Int32Field),
}

// Encode creates a moderation voting packet.
func Encode(ticketID int32, secondsLeft int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(ticketID), codec.Int32(secondsLeft))
}
