// Package joinfailed encodes GAME_CENTER_JOINING_FAILED responses.
package joinfailed

import "github.com/niflaot/pixels/networking/codec"

// Header identifies GAME_CENTER_JOINING_FAILED.
const Header uint16 = 1730

// Encode creates one GAME_CENTER_JOINING_FAILED response.
func Encode(reason int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(reason))
}
