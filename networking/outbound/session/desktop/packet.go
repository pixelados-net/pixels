// Package desktop contains the DESKTOP_VIEW outbound packet.
package desktop

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the DESKTOP_VIEW packet identifier.
	Header uint16 = 122
)

// Definition describes the DESKTOP_VIEW payload fields.
var Definition = codec.Definition{}

// Encode creates a DESKTOP_VIEW packet.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, Definition)
}
