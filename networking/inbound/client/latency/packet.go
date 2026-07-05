// Package latency contains the CLIENT_LATENCY inbound packet.
package latency

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the CLIENT_LATENCY packet identifier.
	Header uint16 = 295
)

// Payload contains the unpacked CLIENT_LATENCY fields.
type Payload struct {
	// RequestID is the requestId protocol field.
	RequestID int32
}

// Definition describes the CLIENT_LATENCY payload fields.
var Definition = codec.Definition{
	codec.Named("requestId", codec.Int32Field),
}

// Decode unpacks a CLIENT_LATENCY packet payload.
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
func payloadFromValues(values []codec.Value) Payload {
	var payload Payload
	payload.RequestID = values[0].Int32

	return payload
}
