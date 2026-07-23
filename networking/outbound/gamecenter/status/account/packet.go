// Package account encodes GAME_CENTER_STATUS responses.
package account

import "github.com/niflaot/pixels/networking/codec"

// Header identifies GAME_CENTER_STATUS.
const Header uint16 = 2893

// Encode creates one GAME_CENTER_STATUS response.
func Encode(gameTypeID int32, freeGamesLeft int32, gamesPlayedTotal int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(gameTypeID), codec.Int32(freeGamesLeft), codec.Int32(gamesPlayedTotal))
}
