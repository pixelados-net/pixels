// Package offer contains the TARGET_OFFER outbound packet.
package offer

import "github.com/niflaot/pixels/networking/codec"

const (
	// Header identifies TARGET_OFFER.
	Header uint16 = 119
)

// Offer contains one targeted offer projection.
type Offer struct {
	// TrackingState stores the client presentation state.
	TrackingState int32
	// ID identifies the offer.
	ID int32
	// Identifier stores its stable key.
	Identifier string
	// ProductCode stores the primary product code.
	ProductCode string
	// PriceCredits stores the credits price.
	PriceCredits int32
	// PricePoints stores the points price.
	PricePoints int32
	// PointsType identifies the points currency.
	PointsType int32
	// PurchaseLimit stores the player limit.
	PurchaseLimit int32
	// ExpirationSeconds stores remaining availability.
	ExpirationSeconds int32
	// Title stores localized display text.
	Title string
	// Description stores localized description text.
	Description string
	// ImageURL stores banner artwork.
	ImageURL string
	// IconURL stores icon artwork.
	IconURL string
	// Type identifies the client layout.
	Type int32
	// SubProducts stores additional product codes.
	SubProducts []string
}

// Encode creates a TARGET_OFFER packet.
func Encode(offer Offer) (codec.Packet, error) {
	definition := codec.Definition{codec.Int32Field, codec.Int32Field, codec.StringField, codec.StringField, codec.Int32Field,
		codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField,
		codec.StringField, codec.StringField, codec.StringField, codec.Int32Field, codec.Int32Field}
	payload, err := codec.AppendPayload(nil, definition, codec.Int32(offer.TrackingState),
		codec.Int32(offer.ID), codec.String(offer.Identifier),
		codec.String(offer.ProductCode), codec.Int32(offer.PriceCredits), codec.Int32(offer.PricePoints),
		codec.Int32(offer.PointsType), codec.Int32(offer.PurchaseLimit), codec.Int32(offer.ExpirationSeconds),
		codec.String(offer.Title), codec.String(offer.Description), codec.String(offer.ImageURL),
		codec.String(offer.IconURL), codec.Int32(offer.Type), codec.Int32(int32(len(offer.SubProducts))))
	for _, product := range offer.SubProducts {
		if err != nil {
			break
		}
		payload, err = codec.AppendPayload(payload, codec.Definition{codec.StringField}, codec.String(product))
	}

	return codec.Packet{Header: Header, Payload: payload}, err
}
