// Package oldchat contains the USER_SETTINGS_OLD_CHAT inbound packet.
package oldchat

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_SETTINGS_OLD_CHAT.
const Header uint16 = 1262

// Definition describes USER_SETTINGS_OLD_CHAT fields.
var Definition = codec.Definition{codec.Named("oldChat", codec.BooleanField)}

// Payload contains decoded old-chat settings.
type Payload struct {
	// OldChat reports whether legacy chat rendering is selected.
	OldChat bool
}

// Decode decodes USER_SETTINGS_OLD_CHAT.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{OldChat: values[0].Boolean}, nil
}
