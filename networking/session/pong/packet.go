// Package pong contains the client-to-server keepalive pong packet.
package pong

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the CLIENT_PONG packet identifier.
	Header uint16 = 2596
)

// Definition describes the CLIENT_PONG payload.
var Definition = codec.Definition{}

// New creates a CLIENT_PONG packet.
func New() codec.Packet {
	return codec.Packet{Header: Header}
}
