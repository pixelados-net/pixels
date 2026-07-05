// Package ok contains the AUTHENTICATED outbound packet.
package ok

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the AUTHENTICATED packet identifier.
	Header uint16 = 2491
)

// Definition describes the AUTHENTICATED payload fields.
var Definition = codec.Definition{}

// Encode creates a AUTHENTICATED packet.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, Definition)
}
