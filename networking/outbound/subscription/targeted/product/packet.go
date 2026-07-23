// Package product contains the PRODUCT_OFFER outbound packet.
package product

import (
	"github.com/niflaot/pixels/networking/codec"
	catalogoffer "github.com/niflaot/pixels/networking/outbound/catalog/offer"
)

const (
	// Header identifies PRODUCT_OFFER.
	Header uint16 = 3388
)

// Encode creates a PRODUCT_OFFER packet.
func Encode(offer catalogoffer.Offer) (codec.Packet, error) {
	payload, err := catalogoffer.AppendPage(nil, offer)
	return codec.Packet{Header: Header, Payload: payload}, err
}
