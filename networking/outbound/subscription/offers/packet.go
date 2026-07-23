// Package offers contains the CLUB_OFFERS outbound packet.
package offers

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies CLUB_OFFERS.
	Header uint16 = 2405
)

// Offer contains one club duration offer.
type Offer struct {
	// ID identifies the offer.
	ID int32
	// Name stores its product name.
	Name string
	// PriceCredits stores the credit price.
	PriceCredits int32
	// PricePoints stores the points price.
	PricePoints int32
	// PointsType identifies the activity-points currency.
	PointsType int32
	// VIP reports whether the offer grants VIP.
	VIP bool
	// Months stores the complete subscription months.
	Months int32
	// ExtraDays stores granted days after complete months.
	ExtraDays int32
	// Giftable reports whether the offer can target another player.
	Giftable bool
	// DaysLeftAfterPurchase stores projected remaining membership days.
	DaysLeftAfterPurchase int32
	// Year stores the projected expiration year.
	Year int32
	// Month stores the projected expiration month.
	Month int32
	// Day stores the projected expiration day.
	Day int32
}

// Encode creates a CLUB_OFFERS packet.
func Encode(offers []Offer) (codec.Packet, error) {
	payload, err := codec.AppendPayload(nil, codec.Definition{codec.Int32Field}, codec.Int32(int32(len(offers))))
	for _, offer := range offers {
		if err != nil {
			break
		}
		payload, err = Append(payload, offer)
	}

	return codec.Packet{Header: Header, Payload: payload}, err
}

// Append appends one Nitro club-offer record.
func Append(payload []byte, offer Offer) ([]byte, error) {
	return codec.AppendPayload(payload, codec.Definition{
		codec.Int32Field, codec.StringField, codec.BooleanField, codec.Int32Field,
		codec.Int32Field, codec.Int32Field, codec.BooleanField, codec.Int32Field,
		codec.Int32Field, codec.BooleanField, codec.Int32Field, codec.Int32Field,
		codec.Int32Field, codec.Int32Field,
	}, codec.Int32(offer.ID), codec.String(offer.Name), codec.Bool(false),
		codec.Int32(offer.PriceCredits), codec.Int32(offer.PricePoints), codec.Int32(offer.PointsType),
		codec.Bool(offer.VIP), codec.Int32(offer.Months), codec.Int32(offer.ExtraDays),
		codec.Bool(offer.Giftable), codec.Int32(offer.DaysLeftAfterPurchase),
		codec.Int32(offer.Year), codec.Int32(offer.Month), codec.Int32(offer.Day))
}
