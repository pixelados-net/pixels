// Package verificationcoderesult encodes the retired phone verification result.
package verificationcoderesult

import "github.com/niflaot/pixels/networking/codec"

// Header identifies PHONE_TRY_VERIFICATION_CODE_RESULT.
const Header uint16 = 91

// Definition describes the header-only packet.
var Definition = codec.Definition{}

// Encode creates one compatibility packet.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, Definition)
}
