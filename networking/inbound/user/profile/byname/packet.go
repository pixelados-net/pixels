// Package byname contains the USER_PROFILE_BY_NAME inbound packet.
package byname

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_PROFILE_BY_NAME.
const Header uint16 = 2249

// Definition describes USER_PROFILE_BY_NAME fields.
var Definition = codec.Definition{codec.Named("username", codec.StringField)}

// Payload contains decoded USER_PROFILE_BY_NAME fields.
type Payload struct {
	// Username identifies the requested public profile.
	Username string
}

// Decode decodes USER_PROFILE_BY_NAME.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{Username: values[0].String}, nil
}
