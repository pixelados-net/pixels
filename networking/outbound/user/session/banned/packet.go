// Package banned contains the USER_BANNED outbound packet.
package banned

import "github.com/niflaot/pixels/networking/codec"

// Header identifies USER_BANNED.
const Header uint16 = 1683

// Definition describes USER_BANNED fields.
var Definition = codec.Definition{codec.Named("message", codec.StringField)}

// Encode creates a USER_BANNED packet with already-localized text.
func Encode(message string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.String(message))
}
