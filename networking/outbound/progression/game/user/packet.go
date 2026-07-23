// Package user encodes USER_GAME_ACHIEVEMENTS compatibility responses.
package user

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_GAME_ACHIEVEMENTS.
const Header uint16 = 2265

// Encode creates the renderer's header-only compatibility response.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{})
}
