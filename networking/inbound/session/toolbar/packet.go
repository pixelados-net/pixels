// Package toolbar contains the CLIENT_TOOLBAR_TOGGLE inbound packet.
package toolbar

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the CLIENT_TOOLBAR_TOGGLE packet identifier.
	Header uint16 = 2313
)

// Payload contains the unpacked CLIENT_TOOLBAR_TOGGLE fields.
type Payload struct{}

// Definition describes the CLIENT_TOOLBAR_TOGGLE payload fields.
var Definition = codec.Definition{}

// Decode unpacks a CLIENT_TOOLBAR_TOGGLE packet payload.
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
