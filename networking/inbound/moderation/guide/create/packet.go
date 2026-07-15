// Package create contains the moderation create inbound packet.
package create

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation create packet.
const Header uint16 = 3338

// Payload contains decoded moderation create fields.
type Payload struct {
	// Topic stores the decoded wire field.
	Topic int32
	// Description stores the decoded wire field.
	Description string
}

// Definition describes moderation create fields.
var Definition = codec.Definition{
	codec.Named("topic", codec.Int32Field),
	codec.Named("description", codec.StringField),
}

// Decode validates and decodes the moderation create packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{
		Topic:       values[0].Int32,
		Description: values[1].String,
	}, nil
}
