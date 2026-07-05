// Package pong contains the CLIENT_PONG inbound packet.
package pong

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the CLIENT_PONG packet identifier.
	Header uint16 = 2596
)

// Payload contains the unpacked CLIENT_PONG fields.
type Payload struct{}

// Definition describes the CLIENT_PONG payload fields.
var Definition = codec.Definition{}

// Decode unpacks a CLIENT_PONG packet payload.
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
