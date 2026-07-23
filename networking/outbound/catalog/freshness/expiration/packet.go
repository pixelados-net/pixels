// Package expiration contains the CATALOG_PAGE_EXPIRATION outbound packet.
package expiration

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies CATALOG_PAGE_EXPIRATION.
	Header uint16 = 2668
)

// Definition describes the packet payload.
var Definition = codec.Definition{codec.Int32Field, codec.Int32Field, codec.StringField}

// Encode creates a CATALOG_PAGE_EXPIRATION packet.
func Encode(pageID int32, secondsLeft int32, pageName string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(pageID), codec.Int32(secondsLeft), codec.String(pageName))
}
