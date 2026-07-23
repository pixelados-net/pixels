// Package message contains the moderation message outbound packet.
package message

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation message packet.
const Header uint16 = 2030

// Definition describes moderation message fields.
var Definition = codec.Definition{
	codec.Named("message", codec.StringField),
	codec.Named("url", codec.StringField),
}

// Encode creates a moderation message packet.
func Encode(message string, url string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.String(message), codec.String(url))
}
