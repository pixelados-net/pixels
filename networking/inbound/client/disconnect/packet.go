// Package disconnect contains the DISCONNECT inbound packet.
package disconnect

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the DISCONNECT packet identifier.
	Header uint16 = 2445
)

// Payload contains the unpacked DISCONNECT fields.
type Payload struct{}

// Definition describes the DISCONNECT payload fields.
var Definition = codec.Definition{}

// Decode unpacks a DISCONNECT packet payload.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}

	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}

	return payloadFromValues(values), nil
}

// payloadFromValues returns a typed payload from decoded values.
func payloadFromValues([]codec.Value) Payload {
	return Payload{}
}
