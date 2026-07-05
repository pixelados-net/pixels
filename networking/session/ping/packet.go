// Package ping contains the server-to-client keepalive ping packet.
package ping

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the CLIENT_PING packet identifier.
	Header uint16 = 3928
)

// Definition describes the CLIENT_PING payload.
var Definition = codec.Definition{}

// New creates a CLIENT_PING packet.
func New() codec.Packet {
	return codec.Packet{Header: Header}
}
