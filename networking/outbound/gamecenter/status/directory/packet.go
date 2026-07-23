// Package directory encodes GAME_CENTER_DIRECTORY_STATUS responses.
package directory

import "github.com/niflaot/pixels/networking/codec"

// Header identifies GAME_CENTER_DIRECTORY_STATUS.
const Header uint16 = 2246

// Encode creates one GAME_CENTER_DIRECTORY_STATUS response.
func Encode(status int32, blockLength int32, gamesPlayed int32, freeGamesLeft int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(status), codec.Int32(blockLength), codec.Int32(gamesPlayed), codec.Int32(freeGamesLeft))
}
