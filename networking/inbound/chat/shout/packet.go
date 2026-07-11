// Package shout contains the UNIT_CHAT_SHOUT inbound packet.
package shout

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies UNIT_CHAT_SHOUT.
	Header uint16 = 2085
)

// Definition describes the shout message and requested bubble style.
var Definition = codec.Definition{codec.Named("message", codec.StringField), codec.Named("styleId", codec.Int32Field)}

// Payload contains a decoded shout request.
type Payload struct {
	// Message stores the submitted chat text.
	Message string
	// StyleID stores the requested bubble style.
	StyleID int32
}

// Decode decodes a UNIT_CHAT_SHOUT packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}

	return Payload{Message: values[0].String, StyleID: values[1].Int32}, nil
}
