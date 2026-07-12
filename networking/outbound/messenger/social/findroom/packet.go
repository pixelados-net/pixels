// Package findroom contains MESSENGER_FIND_FRIENDS.
package findroom

import "github.com/niflaot/pixels/networking/codec"

// Header identifies MESSENGER_FIND_FRIENDS.
const Header uint16 = 1210

// Encode creates MESSENGER_FIND_FRIENDS.
func Encode(success bool) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.BooleanField}, codec.Bool(success))
}
