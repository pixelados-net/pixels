// Package convertedid contains the CONVERTED_ROOM_ID outbound packet.
package convertedid

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CONVERTED_ROOM_ID.
const Header uint16 = 1331

// Definition describes the exact converted room identifier response.
var Definition = codec.Definition{
	codec.Named("globalId", codec.StringField),
	codec.Named("roomId", codec.Int32Field),
}

// Encode creates a CONVERTED_ROOM_ID packet.
func Encode(globalID string, roomID int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.String(globalID), codec.Int32(roomID))
}
