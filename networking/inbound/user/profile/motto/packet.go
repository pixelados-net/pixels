// Package motto contains the USER_MOTTO inbound packet.
package motto

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_MOTTO.
const Header uint16 = 2228

// Definition describes USER_MOTTO fields.
var Definition = codec.Definition{codec.Named("motto", codec.StringField)}

// Payload contains decoded USER_MOTTO fields.
type Payload struct {
	// Motto stores the requested public motto.
	Motto string
}

// Decode decodes USER_MOTTO.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{Motto: values[0].String}, nil
}
