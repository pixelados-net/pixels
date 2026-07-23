// Package attached contains the moderation attached outbound packet.
package attached

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation attached packet.
const Header uint16 = 1591

// Definition describes moderation attached fields.
var Definition = codec.Definition{
	codec.Named("asGuide", codec.BooleanField),
	codec.Named("requesterID", codec.Int32Field),
	codec.Named("requesterName", codec.StringField),
	codec.Named("topic", codec.Int32Field),
}

// Encode creates a moderation attached packet.
func Encode(asGuide bool, requesterID int32, requesterName string, topic int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Bool(asGuide), codec.Int32(requesterID), codec.String(requesterName), codec.Int32(topic))
}
