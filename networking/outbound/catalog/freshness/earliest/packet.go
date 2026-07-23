// Package earliest contains the CATALOG_EARLIEST_EXPIRY outbound packet.
package earliest

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies CATALOG_EARLIEST_EXPIRY.
	Header uint16 = 2515
)

// Definition describes the packet payload.
var Definition = codec.Definition{codec.StringField, codec.Int32Field}

// Encode creates a CATALOG_EARLIEST_EXPIRY packet.
func Encode(pageName string, secondsLeft int32) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.String(pageName), codec.Int32(secondsLeft))
}
