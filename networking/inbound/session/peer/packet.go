// Package peer contains the PEER_USERS_CLASSIFICATION inbound packet.
package peer

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the PEER_USERS_CLASSIFICATION packet identifier.
	Header uint16 = 1160
)

// Payload contains the unpacked PEER_USERS_CLASSIFICATION fields.
type Payload struct{}

// Definition describes the PEER_USERS_CLASSIFICATION payload fields.
var Definition = codec.Definition{}

// Decode unpacks a PEER_USERS_CLASSIFICATION packet payload.
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
