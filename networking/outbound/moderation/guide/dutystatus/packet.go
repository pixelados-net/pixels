// Package dutystatus contains the moderation dutystatus outbound packet.
package dutystatus

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation dutystatus packet.
const Header uint16 = 1548

// Definition describes moderation dutystatus fields.
var Definition = codec.Definition{
	codec.Named("onDuty", codec.BooleanField),
	codec.Named("guides", codec.Int32Field),
	codec.Named("bullies", codec.Int32Field),
	codec.Named("guardians", codec.Int32Field),
}

// Encode creates a moderation dutystatus packet.
func Encode(onDuty bool, guides int32, bullies int32, guardians int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Bool(onDuty), codec.Int32(guides), codec.Int32(bullies), codec.Int32(guardians))
}
