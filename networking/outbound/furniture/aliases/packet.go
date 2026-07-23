// Package aliases encodes the FURNITURE_ALIASES outbound response.
package aliases

import "github.com/niflaot/pixels/networking/codec"

// Header identifies FURNITURE_ALIASES.
const Header uint16 = 1723

// Encode creates an explicit empty furniture alias map.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.Int32Field}, codec.Int32(0))
}
