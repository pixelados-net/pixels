// Package result contains the retired CHANGE_EMAIL_RESULT outbound packet.
package result

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CHANGE_EMAIL_RESULT.
const Header uint16 = 1815

// Definition describes CHANGE_EMAIL_RESULT fields.
var Definition = codec.Definition{codec.Named("result", codec.Int32Field)}

// Encode creates a CHANGE_EMAIL_RESULT packet.
//
// Deprecated: email is CMS-owned and intentionally has no hotel behavior.
func Encode(result int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(result))
}
