// Package settings contains one Nitro social-group inbound packet.
package settings

import "github.com/niflaot/pixels/networking/codec"

// Header identifies this Nitro packet.
const Header uint16 = 2214

// Payload contains decoded protocol fields.
type Payload struct {
	// GroupID stores the decoded protocol field.
	GroupID int64
	// ReadPolicy stores the decoded protocol field.
	ReadPolicy int16
	// PostMessagePolicy stores the decoded protocol field.
	PostMessagePolicy int16
	// PostThreadPolicy stores the decoded protocol field.
	PostThreadPolicy int16
	// ModeratePolicy stores the decoded protocol field.
	ModeratePolicy int16
}

// Decode validates the header and unpacks every field exactly.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field})
	if err != nil {
		return Payload{}, err
	}
	return Payload{GroupID: int64(values[0].Int32), ReadPolicy: int16(values[1].Int32), PostMessagePolicy: int16(values[2].Int32), PostThreadPolicy: int16(values[3].Int32), ModeratePolicy: int16(values[4].Int32)}, nil
}
