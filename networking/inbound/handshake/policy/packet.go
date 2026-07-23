// Package policy contains the CLIENT_POLICY inbound packet.
package policy

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the CLIENT_POLICY packet identifier.
	Header uint16 = 26979
)

// Payload contains the unpacked CLIENT_POLICY fields.
type Payload struct{}

// Definition describes the CLIENT_POLICY payload fields.
var Definition = codec.Definition{}

// Decode unpacks a CLIENT_POLICY packet payload.
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
