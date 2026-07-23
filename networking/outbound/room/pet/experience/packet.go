// Package experience encodes PET_EXPERIENCE.
package experience

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PET_EXPERIENCE.
const Header uint16 = 2156

// Encode creates PET_EXPERIENCE.
func Encode(petID int64, roomIndex int64, gained int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(int32(petID)), codec.Int32(int32(roomIndex)), codec.Int32(gained))
}
