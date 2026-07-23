// Package post contains one Nitro social-group inbound packet.
package post

import "github.com/niflaot/pixels/networking/codec"

// Header identifies this Nitro packet.
const Header uint16 = 3529

// Payload contains decoded protocol fields.
type Payload struct {
	// GroupID stores the decoded protocol field.
	GroupID int64
	// ThreadID stores the decoded protocol field.
	ThreadID int64
	// Subject stores the decoded protocol field.
	Subject string
	// Message stores the decoded protocol field.
	Message string
}

// Decode validates the header and unpacks every field exactly.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.Int32Field, codec.StringField, codec.StringField})
	if err != nil {
		return Payload{}, err
	}
	return Payload{GroupID: int64(values[0].Int32), ThreadID: int64(values[1].Int32), Subject: values[2].String, Message: values[3].String}, nil
}
