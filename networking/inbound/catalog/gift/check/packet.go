// Package check contains the GET_IS_OFFER_GIFTABLE inbound packet.
package check

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies GET_IS_OFFER_GIFTABLE.
	Header uint16 = 1347
)

// Payload contains giftable lookup fields.
type Payload struct {
	// OfferID identifies the offer.
	OfferID int32
}

// Definition describes the payload.
var Definition = codec.Definition{codec.Named("offerId", codec.Int32Field)}

// Decode decodes a giftable lookup.
func Decode(packet codec.Packet) (Payload, error) {
	if packet.Header != Header {
		return Payload{}, codec.ErrUnexpectedHeader
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		return Payload{}, err
	}
	return Payload{OfferID: values[0].Int32}, nil
}
