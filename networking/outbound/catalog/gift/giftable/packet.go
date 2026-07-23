// Package giftable contains the IS_OFFER_GIFTABLE outbound packet.
package giftable

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies IS_OFFER_GIFTABLE.
	Header uint16 = 761
)

// Definition describes the packet payload.
var Definition = codec.Definition{codec.Int32Field, codec.BooleanField}

// Encode creates an IS_OFFER_GIFTABLE packet.
func Encode(offerID int32, allowed bool) (codec.Packet, error) {
	return codec.NewPacket(Header, Definition, codec.Int32(offerID), codec.Bool(allowed))
}
