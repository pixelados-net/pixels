// Package started contains the moderation started outbound packet.
package started

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation started packet.
const Header uint16 = 3209

// Definition describes moderation started fields.
var Definition = codec.Definition{
	codec.Named("partnerID", codec.Int32Field),
	codec.Named("partnerName", codec.StringField),
	codec.Named("partnerLook", codec.StringField),
	codec.Named("topic", codec.Int32Field),
	codec.Named("description", codec.StringField),
	codec.Named("startedAt", codec.StringField),
}

// Encode creates a moderation started packet.
func Encode(partnerID int32, partnerName string, partnerLook string, topic int32, description string, startedAt string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(partnerID), codec.String(partnerName), codec.String(partnerLook), codec.Int32(topic), codec.String(description), codec.String(startedAt))
}
