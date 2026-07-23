// Package forward contains the FORWARD_TO_SOME_ROOM inbound packet.
package forward

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the FORWARD_TO_SOME_ROOM packet identifier.
	Header uint16 = 1703
)

// Payload contains the unpacked FORWARD_TO_SOME_ROOM fields.
type Payload struct {
	// Action identifies the configured navigator forwarding action.
	Action string
}

// Definition describes the FORWARD_TO_SOME_ROOM payload fields.
var Definition = codec.Definition{codec.Named("action", codec.StringField)}

// Decode unpacks a FORWARD_TO_SOME_ROOM packet payload.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}

	return Payload{Action: values[0].String}, nil
}
