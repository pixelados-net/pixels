// Package welcome contains the retired WELCOME_GIFT_CHANGE_EMAIL inbound packet.
package welcome

import "github.com/niflaot/pixels/networking/codec"

// Header identifies WELCOME_GIFT_CHANGE_EMAIL.
const Header uint16 = 66

// Definition describes the WELCOME_GIFT_CHANGE_EMAIL payload.
var Definition = codec.Definition{codec.Named("email", codec.StringField)}

// Payload contains decoded WELCOME_GIFT_CHANGE_EMAIL fields.
type Payload struct {
	// Email stores the ignored CMS-owned address.
	Email string
}

// Decode decodes a WELCOME_GIFT_CHANGE_EMAIL packet.
//
// Deprecated: the welcome-gift journey is intentionally retired.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{Email: values[0].String}, nil
}
