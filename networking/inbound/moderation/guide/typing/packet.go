// Package typing contains the moderation typing inbound packet.
package typing

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation typing packet.
const Header uint16 = 519

// Payload contains decoded moderation typing fields.
type Payload struct {
	// Typing stores the decoded wire field.
	Typing bool
}

// Definition describes moderation typing fields.
var Definition = codec.Definition{
	codec.Named("typing", codec.BooleanField),
}

// Decode validates and decodes the moderation typing packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{
		Typing: values[0].Boolean,
	}, nil
}
