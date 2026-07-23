// Package preferences contains one Nitro social-group inbound packet.
package preferences

import "github.com/niflaot/pixels/networking/codec"

// Header identifies this Nitro packet.
const Header uint16 = 3435

// Payload contains decoded protocol fields.
type Payload struct {
	// GroupID stores the decoded protocol field.
	GroupID int64
	// State stores the decoded protocol field.
	State int16
	// OnlyAdminsDecorate stores the decoded protocol field.
	OnlyAdminsDecorate int32
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
	return Payload{GroupID: int64(values[0].Int32), State: int16(values[1].Int32), OnlyAdminsDecorate: values[2].Int32}, nil
}
