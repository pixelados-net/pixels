// Package ok contains the PURCHASE_OK outbound packet.
package ok

import (
	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/outbound/catalog/offer"
)

const (
	// Header is the PURCHASE_OK packet identifier.
	Header uint16 = 869
)

// Encode creates a PURCHASE_OK packet.
func Encode(value offer.Offer) (codec.Packet, error) {
	payload, err := offer.AppendPurchase(nil, value)
	if err != nil {
		return codec.Packet{}, err
	}

	return codec.Packet{Header: Header, Payload: payload}, nil
}
