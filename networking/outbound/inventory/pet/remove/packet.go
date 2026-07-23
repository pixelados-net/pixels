// Package remove encodes USER_PET_REMOVE.
package remove

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_PET_REMOVE.
const Header uint16 = 3253

// Encode creates USER_PET_REMOVE.
func Encode(petID int64) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(int32(petID)))
}
