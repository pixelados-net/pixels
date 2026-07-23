// Package failure encodes PET_SCRATCH_FAILED.
package failure

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PET_SCRATCH_FAILED.
const Header uint16 = 1130

// Encode creates PET_SCRATCH_FAILED.
func Encode(currentAge int32, requiredAge int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(currentAge), codec.Int32(requiredAge))
}
