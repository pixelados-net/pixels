// Package status encodes PET_STATUS.
package status

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PET_STATUS.
const Header uint16 = 1907

// Encode creates PET_STATUS.
func Encode(roomIndex int64, petID int64, canBreed bool, canHarvest bool, canRevive bool, hasBreedingPermission bool) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{
		codec.Int32Field, codec.Int32Field, codec.BooleanField, codec.BooleanField, codec.BooleanField, codec.BooleanField,
	}, codec.Int32(int32(roomIndex)), codec.Int32(int32(petID)), codec.Bool(canBreed), codec.Bool(canHarvest), codec.Bool(canRevive), codec.Bool(hasBreedingPermission))
}
