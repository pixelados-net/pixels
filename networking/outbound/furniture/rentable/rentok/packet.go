// Package rentok encodes RENTABLE_SPACE_RENT_OK responses.
package rentok

import "github.com/niflaot/pixels/networking/codec"

// Header identifies RENTABLE_SPACE_RENT_OK.
const Header uint16 = 2046

// Definition describes the Unix expiry time.
var Definition = codec.Definition{codec.Named("expiry", codec.Int32Field)}

// Encode creates one successful rent response.
func Encode(expiry int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(expiry))
}
