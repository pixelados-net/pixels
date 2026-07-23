// Package interstitialmessage encodes the retired INTERSTITIAL_MESSAGE packet.
package interstitialmessage

import "github.com/niflaot/pixels/networking/codec"

// Header identifies INTERSTITIAL_MESSAGE.
const Header uint16 = 1808

// Definition describes the legacy availability flag.
var Definition = codec.Definition{codec.Named("available", codec.BooleanField)}

// Encode creates one compatibility packet.
func Encode(available bool) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Bool(available))
}
