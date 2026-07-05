// Package restore contains the RESTORE_CLIENT outbound packet.
package restore

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the RESTORE_CLIENT packet identifier.
	Header uint16 = 426
)

// Definition describes the RESTORE_CLIENT payload fields.
var Definition = codec.Definition{}

// Encode creates a RESTORE_CLIENT packet.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, Definition)
}
