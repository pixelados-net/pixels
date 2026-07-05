// Package cancreate contains the CAN_CREATE_ROOM inbound packet.
package cancreate

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the CAN_CREATE_ROOM packet identifier.
	Header uint16 = 2128
)

// Payload contains the unpacked CAN_CREATE_ROOM fields.
type Payload struct{}

// Definition describes the CAN_CREATE_ROOM payload fields.
var Definition = codec.Definition{}

// Decode unpacks a CAN_CREATE_ROOM packet payload.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	if _, err := codec.DecodePacketExact(packet, Definition); err != nil {
		return Payload{}, err
	}
	return Payload{}, nil
}
