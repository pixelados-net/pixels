// Package visituser contains the VISIT_USER inbound packet.
package visituser

import "github.com/niflaot/pixels/networking/codec"

// Header identifies VISIT_USER.
const Header uint16 = 2970

// Definition describes the exact target username.
var Definition = codec.Definition{codec.Named("username", codec.StringField)}

// Payload contains one user-room visit request.
type Payload struct {
	// Username identifies the requested player.
	Username string
}

// Decode decodes VISIT_USER.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{Username: values[0].String}, nil
}
