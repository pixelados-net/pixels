// Package event contains the EVENT_TRACKER inbound packet.
package event

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the EVENT_TRACKER packet identifier.
	Header uint16 = 3457
)

// Payload contains the unpacked EVENT_TRACKER fields.
type Payload struct{}

// Definition describes the EVENT_TRACKER payload fields.
var Definition = codec.Definition{}

// Decode unpacks a EVENT_TRACKER packet payload.
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
