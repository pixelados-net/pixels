// Package roominfo contains the moderation roominfo inbound packet.
package roominfo

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation roominfo packet.
const Header uint16 = 707

// Payload contains decoded moderation roominfo fields.
type Payload struct {
	// RoomID stores the decoded wire field.
	RoomID int32
}

// Definition describes moderation roominfo fields.
var Definition = codec.Definition{
	codec.Named("roomID", codec.Int32Field),
}

// Decode validates and decodes the moderation roominfo packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{
		RoomID: values[0].Int32,
	}, nil
}
