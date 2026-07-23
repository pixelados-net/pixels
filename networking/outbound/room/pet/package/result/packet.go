// Package result encodes PET_OPEN_PACKAGE_RESULT.
package result

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PET_OPEN_PACKAGE_RESULT.
const Header uint16 = 546

// Encode creates PET_OPEN_PACKAGE_RESULT.
func Encode(objectID int64, nameStatus int32, nameInfo string) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field, codec.Int32Field, codec.StringField}, codec.Int32(int32(objectID)), codec.Int32(nameStatus), codec.String(nameInfo))
}
