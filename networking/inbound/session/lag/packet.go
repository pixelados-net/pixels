// Package lag contains the TRACKING_LAG_WARNING_REPORT inbound packet.
package lag

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the TRACKING_LAG_WARNING_REPORT packet identifier.
	Header uint16 = 3847
)

// Payload contains the unpacked TRACKING_LAG_WARNING_REPORT fields.
type Payload struct{}

// Definition describes the TRACKING_LAG_WARNING_REPORT payload fields.
var Definition = codec.Definition{}

// Decode unpacks a TRACKING_LAG_WARNING_REPORT packet payload.
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
