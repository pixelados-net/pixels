// Package playing encodes PLAYING_GAME responses.
package playing

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PLAYING_GAME.
const Header uint16 = 448

// Encode creates one PLAYING_GAME response.
func Encode(isPlaying bool) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.BooleanField}, codec.Bool(isPlaying))
}
