// Package reply contains the moderation reply outbound packet.
package reply

import "github.com/niflaot/pixels/networking/codec"

// Header identifies the moderation reply packet.
const Header uint16 = 3796

// Definition describes the localized reply.
var Definition = codec.Definition{codec.StringField}

// Encode creates a moderation reply packet.
func Encode(message string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.String(message))
}
