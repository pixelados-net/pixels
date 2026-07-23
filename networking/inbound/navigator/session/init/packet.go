// Package init contains the NAVIGATOR_INIT inbound packet.
package init

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the NAVIGATOR_INIT packet identifier.
	Header uint16 = 2110
)

// Payload contains the unpacked NAVIGATOR_INIT fields.
type Payload struct{}

// Definition describes the NAVIGATOR_INIT payload fields.
var Definition = codec.Definition{}

// Decode unpacks a NAVIGATOR_INIT packet payload.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}

	if _, err := codec.DecodePacketExact(packet, Definition); err != nil {
		return Payload{}, err
	}

	return Payload{}, nil
}
