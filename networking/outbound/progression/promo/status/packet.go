// Package status encodes BADGE_REQUEST_FULFILLED responses.
package status

import "github.com/niflaot/pixels/networking/codec"

// Header identifies BADGE_REQUEST_FULFILLED.
const Header uint16 = 2998

// Encode creates one promotional badge claim status.
func Encode(code string, fulfilled bool) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.StringField, codec.BooleanField}, codec.String(code), codec.Bool(fulfilled))
}
