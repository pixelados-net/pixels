// Package info contains one Nitro social-group inbound packet.
package info

import "github.com/niflaot/pixels/networking/codec"

// Header identifies this Nitro packet.
const Header uint16 = 2991

// Payload contains decoded protocol fields.
type Payload struct {
	// GroupID stores the decoded protocol field.
	GroupID int64
	// Flag stores the decoded protocol field.
	Flag bool
}

// Decode validates the header and unpacks every field exactly.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.BooleanField})
	if err != nil {
		return Payload{}, err
	}
	return Payload{GroupID: int64(values[0].Int32), Flag: values[1].Boolean}, nil
}
