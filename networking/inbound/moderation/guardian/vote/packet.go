// Package vote contains the moderation vote inbound packet.
package vote

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation vote packet.
const Header uint16 = 3961

// Payload contains decoded moderation vote fields.
type Payload struct {
	// Vote stores the decoded wire field.
	Vote int32
}

// Definition describes moderation vote fields.
var Definition = codec.Definition{
	codec.Named("vote", codec.Int32Field),
}

// Decode validates and decodes the moderation vote packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{
		Vote: values[0].Int32,
	}, nil
}
