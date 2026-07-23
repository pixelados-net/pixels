// Package privatechat contains MESSENGER_CHAT.
package privatechat

import "github.com/niflaot/pixels/networking/codec"

// Header identifies MESSENGER_CHAT.
const Header uint16 = 3567

// Payload contains private-chat fields.
type Payload struct {
	// PlayerID identifies the message recipient.
	PlayerID int64
	// Message stores private message text.
	Message string
}

// Decode unpacks MESSENGER_CHAT.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, codec.Definition{codec.Named("playerId", codec.Int32Field), codec.Named("message", codec.StringField)})
	if err != nil {
		return Payload{}, err
	}
	return Payload{PlayerID: int64(values[0].Int32), Message: values[1].String}, nil
}
