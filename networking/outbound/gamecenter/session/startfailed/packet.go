// Package startfailed encodes GAME_CENTER_STARTING_GAME_FAILED responses.
package startfailed

import "github.com/niflaot/pixels/networking/codec"

// Header identifies GAME_CENTER_STARTING_GAME_FAILED.
const Header uint16 = 2142

// Encode creates one GAME_CENTER_STARTING_GAME_FAILED response.
func Encode(reason int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(reason))
}
