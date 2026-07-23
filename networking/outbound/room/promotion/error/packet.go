// Package error encodes ROOM_AD_ERROR responses.
package error

import "github.com/niflaot/pixels/networking/codec"

// Header identifies ROOM_AD_ERROR.
const Header uint16 = 1759

// Definition describes the error code and filtered copy.
var Definition = codec.Definition{codec.Named("errorCode", codec.Int32Field), codec.Named("filteredText", codec.StringField)}

// Encode creates one room-ad error.
func Encode(code int32, filteredText string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(code), codec.String(filteredText))
}
