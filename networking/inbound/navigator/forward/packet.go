// Package forward contains the FORWARD_TO_SOME_ROOM inbound packet.
package forward

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the FORWARD_TO_SOME_ROOM packet identifier.
	Header uint16 = 1703
)

// Payload contains the unpacked FORWARD_TO_SOME_ROOM fields.
type Payload struct{}

// Definition describes the FORWARD_TO_SOME_ROOM payload fields.
var Definition = codec.Definition{}

// Decode unpacks a FORWARD_TO_SOME_ROOM packet payload.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	if _, err := codec.DecodePacketExact(packet, Definition); err != nil {
		return Payload{}, err
	}
	return Payload{}, nil
}
