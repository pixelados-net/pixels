// Package message contains the moderation message inbound packet.
package message

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation message packet.
const Header uint16 = 3899

// Payload contains decoded moderation message fields.
type Payload struct {
	// Message stores the decoded wire field.
	Message string
}

// Definition describes moderation message fields.
var Definition = codec.Definition{
	codec.Named("message", codec.StringField),
}

// Decode validates and decodes the moderation message packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{
		Message: values[0].String,
	}, nil
}
