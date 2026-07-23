// Package change contains the CHANGE_USERNAME inbound packet.
package change

import "github.com/niflaot/pixels/networking/codec"

// Header identifies CHANGE_USERNAME.
const Header uint16 = 2977

// Definition describes CHANGE_USERNAME fields.
var Definition = codec.Definition{codec.Named("username", codec.StringField)}

// Payload contains decoded username change fields.
type Payload struct {
	// Username stores the requested visible name.
	Username string
}

// Decode decodes CHANGE_USERNAME.
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
