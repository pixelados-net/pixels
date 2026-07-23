// Package roomchatlog contains the moderation roomchatlog inbound packet.
package roomchatlog

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation roomchatlog packet.
const Header uint16 = 2587

// Payload contains decoded moderation roomchatlog fields.
type Payload struct {
	// Unused stores the decoded wire field.
	Unused int32
	// RoomID stores the decoded wire field.
	RoomID int32
}

// Definition describes moderation roomchatlog fields.
var Definition = codec.Definition{
	codec.Named("unused", codec.Int32Field),
	codec.Named("roomID", codec.Int32Field),
}

// Decode validates and decodes the moderation roomchatlog packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{
		Unused: values[0].Int32,
		RoomID: values[1].Int32,
	}, nil
}
