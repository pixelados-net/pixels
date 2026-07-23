// Package moderatemessage contains one Nitro social-group inbound packet.
package moderatemessage

import "github.com/niflaot/pixels/networking/codec"

// Header identifies this Nitro packet.
const Header uint16 = 286

// Payload contains decoded protocol fields.
type Payload struct {
	// GroupID stores the decoded protocol field.
	GroupID int64
	// MessageID stores the decoded protocol field.
	MessageID int64
	// State stores the decoded protocol field.
	State int16
	// ThreadID stores the decoded protocol field.
	ThreadID int64
}

// Decode validates the header and unpacks every field exactly.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field})
	if err != nil {
		return Payload{}, err
	}
	return Payload{GroupID: int64(values[0].Int32), MessageID: int64(values[1].Int32), State: int16(values[2].Int32), ThreadID: int64(values[3].Int32)}, nil
}
