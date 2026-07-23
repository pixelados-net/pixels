// Package alert contains the GENERIC_ALERT outbound packet.
package alert

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the GENERIC_ALERT packet identifier.
	Header uint16 = 3801
)

// Definition describes the GENERIC_ALERT payload fields.
var Definition = codec.Definition{
	codec.Named("message", codec.StringField),
}

// Encode creates a GENERIC_ALERT packet.
func Encode(message string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.String(message))
}
