// Package change contains the retired EMAIL_CHANGE inbound packet.
package change

import "github.com/niflaot/pixels/networking/codec"

// Header identifies EMAIL_CHANGE.
const Header uint16 = 3965

// Definition describes the EMAIL_CHANGE payload.
var Definition = codec.Definition{codec.Named("email", codec.StringField)}

// Payload contains decoded EMAIL_CHANGE fields.
type Payload struct {
	// Email stores the ignored CMS-owned address.
	Email string
}

// Decode decodes an EMAIL_CHANGE packet.
//
// Deprecated: email is CMS-owned and intentionally has no hotel behavior.
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
