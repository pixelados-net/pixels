// Package snapshot contains CAMERA_SNAPSHOT.
package snapshot

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CAMERA_SNAPSHOT.
const Header uint16 = 463

// Definition describes the compatibility snapshot.
var Definition = codec.Definition{codec.StringField, codec.Int32Field}

// Encode creates a compatibility snapshot packet.
func Encode(roomType string, roomID int32) (codec.Packet, error) {
	if roomType == "" || roomID <= 0 {
		return codec.Packet{}, codec.ErrInvalidField
	}
	return codec.NewPacket(Header, Definition, codec.String(roomType), codec.Int32(roomID))
}
