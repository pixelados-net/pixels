// Package limited contains the LIMITED_OFFER_APPEARING_NEXT outbound packet.
package limited

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies LIMITED_OFFER_APPEARING_NEXT.
	Header uint16 = 44
)

// Definition describes the packet payload.
var Definition = codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField}

// Encode creates a LIMITED_OFFER_APPEARING_NEXT packet.
func Encode(seconds int32, pageID int32, offerID int32, productType string) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(seconds), codec.Int32(pageID), codec.Int32(offerID), codec.String(productType))
}
