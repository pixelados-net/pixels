// Package firstofday contains the FIRST_LOGIN_OF_DAY outbound packet.
package firstofday

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the FIRST_LOGIN_OF_DAY packet identifier.
	Header uint16 = 793
)

// Definition describes the FIRST_LOGIN_OF_DAY payload fields.
var Definition = codec.Definition{}

// Encode creates a FIRST_LOGIN_OF_DAY packet.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, Definition)
}
