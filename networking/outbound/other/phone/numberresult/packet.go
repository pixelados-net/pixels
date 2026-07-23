// Package numberresult encodes the retired phone-number result.
package numberresult

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PHONE_TRY_NUMBER_RESULT.
const Header uint16 = 800

// Definition describes the header-only packet.
var Definition = codec.Definition{}

// Encode creates one compatibility packet.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, Definition)
}
