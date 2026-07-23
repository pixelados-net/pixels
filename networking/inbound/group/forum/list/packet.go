// Package list contains one Nitro social-group inbound packet.
package list

import "github.com/niflaot/pixels/networking/codec"

// Header identifies this Nitro packet.
const Header uint16 = 436

// Payload contains decoded protocol fields.
type Payload struct {
	// Mode stores the decoded protocol field.
	Mode int32
	// Start stores the decoded protocol field.
	Start int32
	// Amount stores the decoded protocol field.
	Amount int32
}

// Decode validates the header and unpacks every field exactly.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field})
	if err != nil {
		return Payload{}, err
	}
	return Payload{Mode: values[0].Int32, Start: values[1].Int32, Amount: values[2].Int32}, nil
}
