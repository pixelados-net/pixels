// Package requested encodes PET_OPEN_PACKAGE_REQUESTED.
package requested

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PET_OPEN_PACKAGE_REQUESTED.
const Header uint16 = 2380

// Encode creates PET_OPEN_PACKAGE_REQUESTED.
func Encode(objectID int64, figure string) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.StringField}, codec.Int32(int32(objectID)), codec.String(figure))
}
