// Package product contains the GET_PRODUCT_OFFER inbound packet.
package product

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies GET_PRODUCT_OFFER.
	Header uint16 = 2594
)

// Payload contains a catalog offer lookup.
type Payload struct {
	// OfferID identifies the catalog offer.
	OfferID int32
}

// Definition describes the packet payload.
var Definition = codec.Definition{codec.Int32Field}

// Decode decodes a GET_PRODUCT_OFFER packet.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	v, e := codec.DecodePacketExact(packet, Definition)
	if e != nil {
		return Payload{}, e
	}
	return Payload{OfferID: v[0].Int32}, nil
}
