// Package entryinfo contains the ROOM_INFO_OWNER outbound packet.
package entryinfo

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_INFO_OWNER.
	Header uint16 = 749
)

// Definition describes the entered room identity and ownership fields.
var Definition = codec.Definition{
	codec.Named("roomId", codec.Int32Field),
	codec.Named("isOwner", codec.BooleanField),
}

// Encode creates a ROOM_INFO_OWNER packet.
func Encode(roomID int32, isOwner bool) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(roomID), codec.Bool(isOwner))
}
