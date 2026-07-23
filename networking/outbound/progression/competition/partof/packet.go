// Package partof encodes COMPETITION_USER_PART_OF responses.
package partof

import "github.com/niflaot/pixels/networking/codec"

// Header identifies COMPETITION_USER_PART_OF.
const Header uint16 = 3841

// Encode creates one competition membership status.
func Encode(member bool, targetID int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.BooleanField, codec.Int32Field}, codec.Bool(member), codec.Int32(targetID))
}
