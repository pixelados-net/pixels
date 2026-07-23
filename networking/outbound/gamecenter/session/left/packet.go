// Package left encodes GAME_CENTER_USER_LEFT_GAME responses.
package left

import "github.com/niflaot/pixels/networking/codec"

// Header identifies GAME_CENTER_USER_LEFT_GAME.
const Header uint16 = 3138

// Encode creates one GAME_CENTER_USER_LEFT_GAME response.
func Encode(userID int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(userID))
}
