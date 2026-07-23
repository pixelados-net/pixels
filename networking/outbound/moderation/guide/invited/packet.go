// Package invited contains the moderation invited outbound packet.
package invited

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation invited packet.
const Header uint16 = 219

// Definition describes moderation invited fields.
var Definition = codec.Definition{
	codec.Named("roomID", codec.Int32Field),
	codec.Named("roomName", codec.StringField),
}

// Encode creates a moderation invited packet.
func Encode(roomID int32, roomName string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(roomID), codec.String(roomName))
}
