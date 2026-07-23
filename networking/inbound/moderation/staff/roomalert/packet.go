// Package roomalert contains the moderation roomalert inbound packet.
package roomalert

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation roomalert packet.
const Header uint16 = 3842

// Payload contains decoded moderation roomalert fields.
type Payload struct {
	// Action stores Nitro's moderator room-message action.
	Action int32
	// Message stores the decoded wire field.
	Message string
	// Topic stores the decoded wire field.
	Topic string
}

// Definition describes moderation roomalert fields.
var Definition = codec.Definition{
	codec.Named("action", codec.Int32Field),
	codec.Named("message", codec.StringField),
	codec.Named("topic", codec.StringField),
}

// Decode validates and decodes the moderation roomalert packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{
		Action:  values[0].Int32,
		Message: values[1].String,
		Topic:   values[2].String,
	}, nil
}
