// Package info contains the CLUB_GIFT_INFO outbound packet.
package info

import (
	"github.com/niflaot/pixels/networking/codec"
	catalogoffer "github.com/niflaot/pixels/networking/outbound/catalog/offer"
)

const (
	// Header identifies CLUB_GIFT_INFO.
	Header uint16 = 619
)

// Gift contains one selectable club gift.
type Gift struct {
	// Offer stores the catalog offer shape.
	Offer catalogoffer.Offer
	// VIP reports whether the gift requires VIP.
	VIP bool
	// DaysRequired stores the lifetime club days needed for selection.
	DaysRequired int32
	// Selectable reports whether the player satisfies gift requirements.
	Selectable bool
}

// Encode creates a CLUB_GIFT_INFO packet.
func Encode(daysUntilNext int32, available int32, gifts []Gift) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field, codec.Int32Field},
		codec.Int32(daysUntilNext), codec.Int32(available))
	if err == nil {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(gifts))))
	}
	for _, gift := range gifts {
		if err == nil {
			payload, err = catalogoffer.AppendPage(payload, gift.Offer)
		}
	}
	if err == nil {
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(gifts))))
	}
	for _, gift := range gifts {
		if err == nil {
			payload, err = codec.AppendPayload(payload, codec.Definition{
				codec.Int32Field, codec.BooleanField, codec.Int32Field, codec.BooleanField,
			}, codec.Int32(gift.Offer.ID), codec.Bool(gift.VIP),
				codec.Int32(gift.DaysRequired), codec.Bool(gift.Selectable))
		}
	}

	return codec.Packet{Header: Header, Payload: payload}, err
}
