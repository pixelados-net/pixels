// Package extended contains the CLUB_EXTENDED_OFFER outbound packet.
package extended

import (
	"github.com/niflaot/pixels/networking/codec"
	cluboffers "github.com/niflaot/pixels/networking/outbound/subscription/offers"
)

const (
	// Header identifies CLUB_EXTENDED_OFFER.
	Header uint16 = 3964
)

// Encode creates a CLUB_EXTENDED_OFFER packet.
func Encode(offer cluboffers.Offer, originalCredits int32, originalPoints int32, originalPointsType int32, subscriptionDaysLeft int32) (codec.Packet, error) {
	payload, err := cluboffers.Append(nil, offer)
	if err == nil {
		payload, err = codec.AppendPayload(payload, codec.Definition{
			codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field,
		}, codec.Int32(originalCredits), codec.Int32(originalPoints),
			codec.Int32(originalPointsType), codec.Int32(subscriptionDaysLeft))
	}
	return codec.Packet{Header: Header, Payload: payload}, err
}
