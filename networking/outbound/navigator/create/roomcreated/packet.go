// Package roomcreated contains the ROOM_CREATED outbound packet.
package roomcreated

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the ROOM_CREATED packet identifier.
	Header uint16 = 1304
)

// Definition describes the ROOM_CREATED payload fields.
var Definition = codec.Definition{
	codec.Named("roomId", codec.Int32Field),
	codec.Named("roomName", codec.StringField),
}

// Encode creates a ROOM_CREATED packet.
func Encode(roomID int32, roomName string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(roomID), codec.String(roomName))
}
