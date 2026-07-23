// Package cancelled encodes QUEST_CANCELLED responses.
package cancelled

import "github.com/niflaot/pixels/networking/codec"

// Header identifies QUEST_CANCELLED.
const Header uint16 = 3027

// Encode creates one quest cancellation response.
func Encode(expired bool) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.BooleanField}, codec.Bool(expired))
}
