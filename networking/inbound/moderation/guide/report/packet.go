// Package report contains the moderation report inbound packet.
package report

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation report packet.
const Header uint16 = 3969

// Payload contains decoded moderation report fields.
type Payload struct {
	// Reason stores the decoded wire field.
	Reason string
}

// Definition describes moderation report fields.
var Definition = codec.Definition{
	codec.Named("reason", codec.StringField),
}

// Decode validates and decodes the moderation report packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{
		Reason: values[0].String,
	}, nil
}
