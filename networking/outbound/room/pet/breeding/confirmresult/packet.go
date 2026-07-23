// Package confirmresult encodes PET_CONFIRM_BREEDING_RESULT.
package confirmresult

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PET_CONFIRM_BREEDING_RESULT.
const Header uint16 = 1625

// Encode creates PET_CONFIRM_BREEDING_RESULT.
func Encode(nestItemID int64, resultCode int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(int32(nestItemID)), codec.Int32(resultCode))
}
