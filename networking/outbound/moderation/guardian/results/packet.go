// Package results contains the moderation results outbound packet.
package results

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation results packet.
const Header uint16 = 3276

// Definition describes moderation results fields.
var Definition = codec.Definition{
	codec.Named("ticketID", codec.Int32Field),
	codec.Named("result", codec.Int32Field),
	codec.Named("votes", codec.Int32Field),
	codec.Named("total", codec.Int32Field),
}

// Encode creates a moderation results packet.
func Encode(ticketID int32, result int32, votes int32, total int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(ticketID), codec.Int32(result), codec.Int32(votes), codec.Int32(total))
}
