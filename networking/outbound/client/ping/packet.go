// Package ping contains the CLIENT_PING outbound packet.
package ping

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the CLIENT_PING packet identifier.
	Header uint16 = 3928
)

// Definition describes the CLIENT_PING payload fields.
var Definition = codec.Definition{}

// Encode creates a CLIENT_PING packet.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, Definition)
}
