// Package latency contains the CLIENT_LATENCY_MEASURE inbound packet.
package latency

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the CLIENT_LATENCY_MEASURE packet identifier.
	Header uint16 = 96
)

// Payload contains the unpacked CLIENT_LATENCY_MEASURE fields.
type Payload struct{}

// Definition describes the CLIENT_LATENCY_MEASURE payload fields.
var Definition = codec.Definition{}

// Decode unpacks a CLIENT_LATENCY_MEASURE packet payload.
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
