// Package unbanned contains the ROOM_BAN_REMOVE outbound packet.
package unbanned

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies ROOM_BAN_REMOVE.
	Header uint16 = 3429
)

// Definition describes ROOM_BAN_REMOVE fields.
var Definition = codec.Definition{codec.Named("roomId", codec.Int32Field), codec.Named("playerId", codec.Int32Field)}

// Encode creates a ROOM_BAN_REMOVE packet.
func Encode(roomID int32, playerID int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(roomID), codec.Int32(playerID))
}
