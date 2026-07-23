// Package status contains the moderation status inbound packet.
package status

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation status packet.
const Header uint16 = 2746

// Payload contains decoded moderation status fields.
type Payload struct {
	// Enabled stores the decoded wire field.
	Enabled bool
}

// Definition describes moderation status fields.
var Definition = codec.Definition{
	codec.Named("enabled", codec.BooleanField),
}

// Decode validates and decodes the moderation status packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{
		Enabled: values[0].Boolean,
	}, nil
}
