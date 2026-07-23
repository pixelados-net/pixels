// Package decide contains the moderation decide inbound packet.
package decide

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation decide packet.
const Header uint16 = 3365

// Payload contains decoded moderation decide fields.
type Payload struct {
	// Accepted stores the decoded wire field.
	Accepted bool
}

// Definition describes moderation decide fields.
var Definition = codec.Definition{
	codec.Named("accepted", codec.BooleanField),
}

// Decode validates and decodes the moderation decide packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{
		Accepted: values[0].Boolean,
	}, nil
}
