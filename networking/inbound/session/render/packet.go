// Package render contains the RENDER_ROOM inbound packet.
package render

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the RENDER_ROOM packet identifier.
	Header uint16 = 3226
)

// Payload contains the unpacked RENDER_ROOM fields.
type Payload struct{}

// Definition describes the RENDER_ROOM payload fields.
var Definition = codec.Definition{}

// Decode unpacks a RENDER_ROOM packet payload.
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
