// Package typing contains the UNIT_TYPING outbound packet.
package typing

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies UNIT_TYPING.
	Header uint16 = 1717
)

// Definition describes one unit typing-state update.
var Definition = codec.Definition{codec.Named("unitId", codec.Int32Field), codec.Named("typing", codec.Int32Field)}

// Encode creates a UNIT_TYPING packet.
func Encode(unitID int32, active bool) (codec.Packet, error) {
	typing := int32(0)
	if active {
		typing = 1
	}
	return codec.NewPacket(Header, Definition, codec.Int32(unitID), codec.Int32(typing))
}
