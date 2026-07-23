// Package feedback contains the moderation feedback inbound packet.
package feedback

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation feedback packet.
const Header uint16 = 477

// Payload contains decoded moderation feedback fields.
type Payload struct {
	// Recommended stores the decoded wire field.
	Recommended bool
}

// Definition describes moderation feedback fields.
var Definition = codec.Definition{
	codec.Named("recommended", codec.BooleanField),
}

// Decode validates and decodes the moderation feedback packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{
		Recommended: values[0].Boolean,
	}, nil
}
