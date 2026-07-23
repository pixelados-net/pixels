// Package soldout contains the LIMITED_SOLD_OUT outbound packet.
package soldout

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the LIMITED_SOLD_OUT packet identifier.
	Header uint16 = 377
)

// Encode creates a LIMITED_SOLD_OUT packet.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{})
}
