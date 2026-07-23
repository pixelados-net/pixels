// Package score encodes USER_ACHIEVEMENT_SCORE responses.
package score

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_ACHIEVEMENT_SCORE.
const Header uint16 = 1968

// Encode creates one achievement score response.
func Encode(value int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(value))
}
