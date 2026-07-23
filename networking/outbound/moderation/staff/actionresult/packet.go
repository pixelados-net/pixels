// Package actionresult contains the moderation actionresult outbound packet.
package actionresult

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation actionresult packet.
const Header uint16 = 2335

// Definition describes moderation actionresult fields.
var Definition = codec.Definition{
	codec.Named("userID", codec.Int32Field),
	codec.Named("success", codec.BooleanField),
}

// Encode creates a moderation actionresult packet.
func Encode(userID int32, success bool) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(userID), codec.Bool(success))
}
