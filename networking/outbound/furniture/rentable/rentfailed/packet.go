// Package rentfailed encodes RENTABLE_SPACE_RENT_FAILED responses.
package rentfailed

import "github.com/niflaot/pixels/networking/codec"

// Header identifies RENTABLE_SPACE_RENT_FAILED.
const Header uint16 = 1868

// Definition describes the failure reason.
var Definition = codec.Definition{codec.Named("reason", codec.Int32Field)}

// Encode creates one failed rent response.
func Encode(reason int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(reason))
}
