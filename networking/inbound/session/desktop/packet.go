// Package desktop contains the DESKTOP_VIEW inbound packet.
package desktop

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the DESKTOP_VIEW packet identifier.
	Header uint16 = 105
)

// Payload contains the unpacked DESKTOP_VIEW fields.
type Payload struct{}

// Definition describes the DESKTOP_VIEW payload fields.
var Definition = codec.Definition{}

// Decode unpacks a DESKTOP_VIEW packet payload.
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
