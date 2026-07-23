// Package motd contains the MOTD_MESSAGES outbound packet.
package motd

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the MOTD_MESSAGES packet identifier.
	Header uint16 = 2035
)

// Definition describes the MOTD_MESSAGES payload fields.
var Definition = codec.Definition{}

// Encode creates a MOTD_MESSAGES packet.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, Definition)
}
