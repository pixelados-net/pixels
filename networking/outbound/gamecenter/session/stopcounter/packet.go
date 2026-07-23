// Package stopcounter encodes GAME_CENTER_STOP_COUNTER responses.
package stopcounter

import "github.com/niflaot/pixels/networking/codec"

// Header identifies GAME_CENTER_STOP_COUNTER.
const Header uint16 = 3191

// Encode creates one header-only GAME_CENTER_STOP_COUNTER response.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{})
}
