// Package favoritechanged contains the USER_FAVORITE_ROOM outbound packet.
package favoritechanged

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the USER_FAVORITE_ROOM packet identifier.
	Header uint16 = 2524
)

// Definition describes the USER_FAVORITE_ROOM payload fields.
var Definition = codec.Definition{
	codec.Named("roomId", codec.Int32Field),
	codec.Named("added", codec.BooleanField),
}

// Encode creates a USER_FAVORITE_ROOM packet.
func Encode(roomID int32, added bool) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(roomID), codec.Bool(added))
}
