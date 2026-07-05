// Package performance contains the TRACKING_PERFORMANCE_LOG inbound packet.
package performance

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the TRACKING_PERFORMANCE_LOG packet identifier.
	Header uint16 = 3230
)

// Payload contains the unpacked TRACKING_PERFORMANCE_LOG fields.
type Payload struct{}

// Definition describes the TRACKING_PERFORMANCE_LOG payload fields.
var Definition = codec.Definition{}

// Decode unpacks a TRACKING_PERFORMANCE_LOG packet payload.
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
