// Package updated contains the CATALOG_PUBLISHED outbound packet.
package updated

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header is the CATALOG_PUBLISHED packet identifier.
	Header uint16 = 1866
)

// Definition describes the CATALOG_PUBLISHED payload fields.
var Definition = codec.Definition{codec.Named("unknown", codec.BooleanField)}

// Encode creates a CATALOG_PUBLISHED packet.
func Encode() (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Bool(false))
}
