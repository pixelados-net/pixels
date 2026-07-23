// Package check contains the CHECK_USERNAME inbound packet.
package check

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CHECK_USERNAME.
const Header uint16 = 3950

// Definition describes CHECK_USERNAME fields.
var Definition = codec.Definition{codec.Named("username", codec.StringField)}

// Payload contains decoded username check fields.
type Payload struct {
	// Username stores the requested visible name.
	Username string
}

// Decode decodes CHECK_USERNAME.
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
