// Package status contains the retired EMAIL_STATUS outbound packet.
package status

import "github.com/niflaot/pixels/networking/codec"

// Header identifies EMAIL_STATUS.
const Header uint16 = 612

// Definition describes EMAIL_STATUS fields.
var Definition = codec.Definition{
	codec.Named("email", codec.StringField),
	codec.Named("verified", codec.BooleanField),
	codec.Named("allowChange", codec.BooleanField),
}

// Encode creates an EMAIL_STATUS packet.
//
// Deprecated: email is CMS-owned and intentionally has no hotel behavior.
func Encode(email string, verified bool, allowChange bool) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.String(email), codec.Bool(verified), codec.Bool(allowChange))
}
