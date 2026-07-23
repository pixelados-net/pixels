// Package settings encodes the superseded ROOM_SETTINGS_ERROR packet.
package settings

import "github.com/niflaot/pixels/networking/codec"

// Header identifies ROOM_SETTINGS_ERROR.
const Header uint16 = 2897

// Definition describes the superseded room and error code.
var Definition = codec.Definition{codec.Named("roomId", codec.Int32Field), codec.Named("code", codec.Int32Field)}

// Encode creates a compatibility ROOM_SETTINGS_ERROR packet.
func Encode(roomID int32, code int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(roomID), codec.Int32(code))
}
