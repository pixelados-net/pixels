// Package result contains the moderation result outbound packet.
package result

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation result packet.
const Header uint16 = 3635

// Definition describes result code and localized message.
var Definition = codec.Definition{codec.Int32Field, codec.StringField}

// Encode creates a moderation result packet.
func Encode(code int32, message string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(code), codec.String(message))
}
