// Package homeroom contains the USER_HOME_ROOM outbound packet.
package homeroom

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_HOME_ROOM.
const Header uint16 = 2875

// Definition describes USER_HOME_ROOM fields.
var Definition = codec.Definition{codec.Named("homeRoomId", codec.Int32Field), codec.Named("roomIdToEnter", codec.Int32Field)}

// Encode creates a USER_HOME_ROOM packet.
func Encode(homeRoomID int32, roomIDToEnter int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(homeRoomID), codec.Int32(roomIDToEnter))
}
