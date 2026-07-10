// Package offer encodes catalog offers shared by catalog response packets.
package offer

import "github.com/niflaot/pixels/networking/codec"

// Product describes one product inside a catalog offer.
type Product struct {
	// Type identifies floor, wall, badge, or effect products.
	Type string
	// ClassID identifies the client rendering class.
	ClassID int32
	// ExtraData stores initial product state.
	ExtraData string
	// Amount stores the number of granted products.
	Amount int32
	// Limited reports whether the product has numbered stock.
	Limited bool
	// LimitedStack stores total numbered stock.
	LimitedStack int32
	// LimitedRemaining stores available numbered stock.
	LimitedRemaining int32
}

// Offer describes one client-visible catalog offer.
type Offer struct {
	// ID identifies the purchasable catalog offer.
	ID int32
	// LocalizationID identifies client product text.
	LocalizationID string
	// Rent reports whether the offer is rented.
	Rent bool
	// CostCredits stores the credits price.
	CostCredits int32
	// CostPoints stores the activity-points price.
	CostPoints int32
	// PointsType identifies the activity-points currency.
	PointsType int32
	// Giftable reports whether the offer can be gifted.
	Giftable bool
	// Products stores products granted by the offer.
	Products []Product
	// ClubLevel stores the required club level.
	ClubLevel int32
	// BundlePurchaseAllowed reports whether bulk purchase is allowed.
	BundlePurchaseAllowed bool
	// Pet reports whether the offer creates a pet.
	Pet bool
	// PreviewImage stores an optional client preview image.
	PreviewImage string
}

// AppendPage appends an offer in CATALOG_PAGE response shape.
func AppendPage(dst []byte, value Offer) ([]byte, error) {
	dst, err := appendBase(dst, value)
	if err != nil {
		return dst, err
	}

	return codec.AppendPayload(dst, codec.Definition{
		codec.BooleanField, codec.BooleanField, codec.StringField,
	}, codec.Bool(value.BundlePurchaseAllowed), codec.Bool(value.Pet), codec.String(value.PreviewImage))
}

// AppendPurchase appends an offer in PURCHASE_OK response shape.
func AppendPurchase(dst []byte, value Offer) ([]byte, error) {
	dst, err := appendBase(dst, value)
	if err != nil {
		return dst, err
	}

	return codec.AppendPayload(dst, codec.Definition{codec.BooleanField}, codec.Bool(value.BundlePurchaseAllowed))
}

// appendBase appends fields common to catalog page and purchase responses.
func appendBase(dst []byte, value Offer) ([]byte, error) {
	dst, err := codec.AppendPayload(dst, codec.Definition{
		codec.Int32Field, codec.StringField, codec.BooleanField, codec.Int32Field,
		codec.Int32Field, codec.Int32Field, codec.BooleanField, codec.Int32Field,
	}, codec.Int32(value.ID), codec.String(value.LocalizationID), codec.Bool(value.Rent),
		codec.Int32(value.CostCredits), codec.Int32(value.CostPoints), codec.Int32(value.PointsType),
		codec.Bool(value.Giftable), codec.Int32(int32(len(value.Products))))
	if err != nil {
		return dst, err
	}
	for _, product := range value.Products {
		dst, err = appendProduct(dst, product)
		if err != nil {
			return dst, err
		}
	}

	return codec.AppendPayload(dst, codec.Definition{codec.Int32Field}, codec.Int32(value.ClubLevel))
}

// appendProduct appends one catalog product.
func appendProduct(dst []byte, product Product) ([]byte, error) {
	dst, err := codec.AppendPayload(dst, codec.Definition{
		codec.StringField, codec.Int32Field, codec.StringField, codec.Int32Field, codec.BooleanField,
	}, codec.String(product.Type), codec.Int32(product.ClassID), codec.String(product.ExtraData),
		codec.Int32(product.Amount), codec.Bool(product.Limited))
	if err != nil || !product.Limited {
		return dst, err
	}

	return codec.AppendPayload(dst, codec.Definition{codec.Int32Field, codec.Int32Field},
		codec.Int32(product.LimitedStack), codec.Int32(product.LimitedRemaining))
}
