// Package seasonal contains the SEASONAL_CALENDAR_OFFER outbound packet.
package seasonal

import (
	"github.com/niflaot/pixels/networking/codec"
	catalogoffer "github.com/niflaot/pixels/networking/outbound/catalog/offer"
)

const (
	// Header identifies SEASONAL_CALENDAR_OFFER.
	Header uint16 = 1889
)

// Encode creates a SEASONAL_CALENDAR_OFFER packet.
func Encode(pageID int32, offer catalogoffer.Offer) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(pageID))
	if err == nil {
		payload, err = catalogoffer.AppendPage(payload, offer)
	}

	return codec.Packet{Header: Header, Payload: payload}, err
}
