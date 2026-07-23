// Package supplemented encodes PET_SUPPLEMENT.
package supplemented

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PET_SUPPLEMENT.
const Header uint16 = 3441

// Encode creates PET_SUPPLEMENT.
func Encode(petID int64, userID int64, supplementType int32) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(int32(petID)), codec.Int32(int32(userID)), codec.Int32(supplementType))
}
