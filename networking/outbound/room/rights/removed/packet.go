// Package removed contains the ROOM_RIGHTS_LIST_REMOVE outbound packet.
package removed

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_RIGHTS_LIST_REMOVE.
	Header uint16 = 1327
)

// Definition describes ROOM_RIGHTS_LIST_REMOVE fields.
var Definition = codec.Definition{codec.Named("roomId", codec.Int32Field), codec.Named("playerId", codec.Int32Field)}

// Encode creates a ROOM_RIGHTS_LIST_REMOVE packet.
func Encode(roomID int32, playerID int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(roomID), codec.Int32(playerID))
}
