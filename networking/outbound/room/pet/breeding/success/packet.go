// Package success encodes PET_NEST_BREEDING_SUCCESS.
package success

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PET_NEST_BREEDING_SUCCESS.
const Header uint16 = 2527

// Encode creates PET_NEST_BREEDING_SUCCESS.
func Encode(petID int64, rarityCategory int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(int32(petID)), codec.Int32(rarityCategory))
}
