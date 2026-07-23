// Package updatethread contains one Nitro social-group inbound packet.
package updatethread

import "github.com/niflaot/pixels/networking/codec"

// Header identifies this Nitro packet.
const Header uint16 = 3045

// Payload contains decoded protocol fields.
type Payload struct {
	// GroupID stores the decoded protocol field.
	GroupID int64
	// ThreadID stores the decoded protocol field.
	ThreadID int64
	// Locked stores the decoded protocol field.
	Locked bool
	// Pinned stores the decoded protocol field.
	Pinned bool
}

// Decode validates the header and unpacks every field exactly.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.Int32Field, codec.BooleanField, codec.BooleanField})
	if err != nil {
		return Payload{}, err
	}
	return Payload{GroupID: int64(values[0].Int32), ThreadID: int64(values[1].Int32), Locked: values[2].Boolean, Pinned: values[3].Boolean}, nil
}
