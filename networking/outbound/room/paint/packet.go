// Package paint encodes a room plane appearance change.
package paint

import "github.com/niflaot/pixels/networking/codec"

// Header is the ROOM_PAINT identifier.
const Header uint16 = 2454

// Encode creates a ROOM_PAINT packet.
func Encode(surface string, value string) (codec.Packet, error) {
	return codec.NewPacket(Header, codec.Definition{codec.StringField, codec.StringField}, codec.String(surface), codec.String(value))
}
