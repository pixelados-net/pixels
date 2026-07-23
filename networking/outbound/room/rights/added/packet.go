// Package added contains the ROOM_RIGHTS_LIST_ADD outbound packet.
package added

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_RIGHTS_LIST_ADD.
	Header uint16 = 2088
)

// Definition describes ROOM_RIGHTS_LIST_ADD fields.
var Definition = codec.Definition{
	codec.Named("roomId", codec.Int32Field),
	codec.Named("playerId", codec.Int32Field),
	codec.Named("username", codec.StringField),
}

// Encode creates a ROOM_RIGHTS_LIST_ADD packet.
func Encode(roomID int32, playerID int32, username string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(roomID), codec.Int32(playerID), codec.String(username))
}
