// Package furniture contains the REQUEST_FURNITURE_INVENTORY inbound packet.
package furniture

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the REQUEST_FURNITURE_INVENTORY packet identifier.
	Header uint16 = 3150
)

// Payload contains the unpacked REQUEST_FURNITURE_INVENTORY fields.
type Payload struct{}

// Definition describes the REQUEST_FURNITURE_INVENTORY payload fields.
var Definition = codec.Definition{}

// Decode unpacks a REQUEST_FURNITURE_INVENTORY packet payload.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	if _, err := codec.DecodePacketExact(packet, Definition); err != nil {
		return Payload{}, err
	}

	return Payload{}, nil
}
