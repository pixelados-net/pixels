// Package duty contains the moderation duty inbound packet.
package duty

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation duty packet.
const Header uint16 = 1922

// Payload contains decoded moderation duty fields.
type Payload struct {
	// Guide stores the decoded wire field.
	Guide bool
	// Bully stores the decoded wire field.
	Bully bool
	// Guardian stores the decoded wire field.
	Guardian bool
	// Enabled stores the decoded wire field.
	Enabled bool
}

// Definition describes moderation duty fields.
var Definition = codec.Definition{
	codec.Named("guide", codec.BooleanField),
	codec.Named("bully", codec.BooleanField),
	codec.Named("guardian", codec.BooleanField),
	codec.Named("enabled", codec.BooleanField),
}

// Decode validates and decodes the moderation duty packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{
		Guide:    values[0].Boolean,
		Bully:    values[1].Boolean,
		Guardian: values[2].Boolean,
		Enabled:  values[3].Boolean,
	}, nil
}
