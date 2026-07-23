// Package search contains one Nitro social-group inbound packet.
package search

import "github.com/niflaot/pixels/networking/codec"

// Header identifies this Nitro packet.
const Header uint16 = 312

// Payload contains decoded protocol fields.
type Payload struct {
	// GroupID stores the decoded protocol field.
	GroupID int64
	// Page stores the decoded protocol field.
	Page int32
	// Query stores the decoded protocol field.
	Query string
	// Level stores the decoded protocol field.
	Level int32
}

// Decode validates the header and unpacks every field exactly.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field})
	if err != nil {
		return Payload{}, err
	}
	return Payload{GroupID: int64(values[0].Int32), Page: values[1].Int32, Query: values[2].String, Level: values[3].Int32}, nil
}
