// Package welcomeresult contains the retired WELCOME_GIFT_CHANGE_EMAIL_RESULT packet.
package welcomeresult

import "github.com/niflaot/pixels/networking/codec"

// Header identifies WELCOME_GIFT_CHANGE_EMAIL_RESULT.
const Header uint16 = 2293

// Definition describes the result code.
var Definition = codec.Definition{codec.Named("result", codec.Int32Field)}

// Encode creates a retired welcome email result packet.
//
// Deprecated: the legacy welcome-gift journey is intentionally retired.
func Encode(result int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(result))
}
