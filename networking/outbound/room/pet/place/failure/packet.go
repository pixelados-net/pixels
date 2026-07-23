// Package failure encodes PET_PLACING_ERROR.
package failure

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PET_PLACING_ERROR.
const Header uint16 = 2913

// Encode creates PET_PLACING_ERROR.
func Encode(code int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(code))
}
