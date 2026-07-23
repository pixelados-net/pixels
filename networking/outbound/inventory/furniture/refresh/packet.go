// Package refresh contains the FURNITURE_INVENTORY_REFRESH outbound packet.
package refresh

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the FURNITURE_INVENTORY_REFRESH packet identifier.
	Header uint16 = 3151
)

// Definition describes the FURNITURE_INVENTORY_REFRESH payload fields.
var Definition = codec.Definition{}

// Encode creates a FURNITURE_INVENTORY_REFRESH packet.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, Definition)
}
