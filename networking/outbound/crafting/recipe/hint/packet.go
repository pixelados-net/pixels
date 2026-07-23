// Package hint contains the CRAFTING_RECIPES_AVAILABLE outbound packet.
package hint

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CRAFTING_RECIPES_AVAILABLE.
const Header uint16 = 2124

// Encode creates one undiscovered recipe match hint packet.
func Encode(count int32, hasExactMatch bool) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.BooleanField}, codec.Int32(count), codec.Bool(hasExactMatch))
}
