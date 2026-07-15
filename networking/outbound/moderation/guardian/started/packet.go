// Package started contains the moderation started outbound packet.
package started

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation started packet.
const Header uint16 = 143

// Definition describes moderation started fields.
var Definition = codec.Definition{
	codec.Named("ticketID", codec.Int32Field),
	codec.Named("chatlog", codec.StringField),
}

// Encode creates a moderation started packet.
func Encode(ticketID int32, chatlog string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(ticketID), codec.String(chatlog))
}
