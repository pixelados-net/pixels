// Package save contains one Nitro social-group inbound packet.
package save

import "github.com/niflaot/pixels/networking/codec"

// Header identifies this Nitro packet.
const Header uint16 = 3137

// Payload contains decoded protocol fields.
type Payload struct {
	// GroupID stores the decoded protocol field.
	GroupID int64
	// Name stores the decoded protocol field.
	Name string
	// Description stores the decoded protocol field.
	Description string
}

// Decode validates the header and unpacks every field exactly.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.StringField, codec.StringField})
	if err != nil {
		return Payload{}, err
	}
	return Payload{GroupID: int64(values[0].Int32), Name: values[1].String, Description: values[2].String}, nil
}
