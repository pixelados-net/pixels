// Package state encodes PET_BREEDING.
package state

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PET_BREEDING.
const Header uint16 = 1746

// Encode creates PET_BREEDING.
func Encode(value int32, ownPetID int64, otherPetID int64) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(value), codec.Int32(int32(ownPetID)), codec.Int32(int32(otherPetID)))
}
