// Package init contains the HANDSHAKE_INIT_DIFFIE inbound packet.
package init

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the HANDSHAKE_INIT_DIFFIE packet identifier.
	Header uint16 = 3110
)

// Payload contains the unpacked HANDSHAKE_INIT_DIFFIE fields.
type Payload struct{}

// Definition describes the HANDSHAKE_INIT_DIFFIE payload fields.
var Definition = codec.Definition{}

// Decode unpacks a HANDSHAKE_INIT_DIFFIE packet payload.
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
