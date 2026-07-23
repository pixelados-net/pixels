// Package selected contains the CLUB_GIFT_SELECTED outbound packet.
package selected

import (
	"github.com/niflaot/pixels/networking/codec"
	catalogoffer "github.com/niflaot/pixels/networking/outbound/catalog/offer"
)

const (
	// Header identifies CLUB_GIFT_SELECTED.
	Header uint16 = 659
)

// Encode creates a CLUB_GIFT_SELECTED packet.
func Encode(productCode string, products []catalogoffer.Product) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.StringField, codec.Int32Field},
		codec.String(productCode), codec.Int32(int32(len(products))))
	for _, product := range products {
		if err != nil {
			break
		}
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField, codec.Int32Field, codec.StringField,
			codec.Int32Field, codec.BooleanField}, codec.String(product.Type), codec.Int32(product.ClassID),
			codec.String(product.ExtraData), codec.Int32(product.Amount), codec.Bool(product.Limited))
		if err == nil && product.Limited {
			payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field, codec.Int32Field},
				codec.Int32(product.LimitedStack), codec.Int32(product.LimitedRemaining))
		}
	}

	return codec.Packet{Header: Header, Payload: payload}, err
}
