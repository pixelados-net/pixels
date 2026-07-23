// Package level encodes PET_LEVEL_UPDATE.
package level

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PET_LEVEL_UPDATE.
const Header uint16 = 2824

// Encode creates PET_LEVEL_UPDATE.
func Encode(roomIndex int64, petID int64, level int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(int32(roomIndex)), codec.Int32(int32(petID)), codec.Int32(level))
}
