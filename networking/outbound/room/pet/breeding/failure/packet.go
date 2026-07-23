// Package failure encodes PET_GO_TO_BREEDING_NEST_FAILURE.
package failure

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PET_GO_TO_BREEDING_NEST_FAILURE.
const Header uint16 = 2621

// Encode creates PET_GO_TO_BREEDING_NEST_FAILURE.
func Encode(reason int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(reason))
}
