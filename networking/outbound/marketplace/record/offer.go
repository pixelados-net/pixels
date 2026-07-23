// Package record encodes shared Marketplace wire records.
package record

import (
	marketcore "github.com/niflaot/pixels/internal/realm/marketplace/core"
	marketrecord "github.com/niflaot/pixels/internal/realm/marketplace/record"
	"github.com/niflaot/pixels/networking/codec"
)

// AppendOffer appends one Nitro Marketplace listing record.
func AppendOffer(dst []byte, offer marketcore.Offer, own bool) ([]byte, error) {
	kind := int32(1)
	if offer.Definition.Kind == "wall" {
		kind = 2
	}
	if offer.Item.LimitedEditionNumber != nil {
		kind = 3
	}
	state := int32(offer.Listing.State) + 1
	dst, err := codec.AppendPayload(dst, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(int32(offer.Listing.ID)), codec.Int32(state), codec.Int32(kind), codec.Int32(int32(offer.Definition.SpriteID)))
	if err != nil {
		return nil, err
	}
	switch kind {
	case 2:
		dst, err = codec.AppendPayload(dst, codec.Definition{codec.StringField}, codec.String(offer.Item.ExtraData))
	case 3:
		dst, err = codec.AppendPayload(dst, codec.Definition{codec.Int32Field, codec.Int32Field}, codec.Int32(*offer.Item.LimitedEditionNumber), codec.Int32(0))
	default:
		dst, err = codec.AppendPayload(dst, codec.Definition{codec.Int32Field, codec.StringField}, codec.Int32(0), codec.String(offer.Item.ExtraData))
	}
	if err != nil {
		return nil, err
	}
	if own {
		minutes := offer.MinutesRemaining
		if offer.Listing.State != marketrecord.StateOpen {
			minutes = 0
		}
		return codec.AppendPayload(dst, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(int32(offer.Listing.RawPrice)), codec.Int32(minutes), codec.Int32(0))
	}
	return codec.AppendPayload(dst, codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}, codec.Int32(int32(offer.BuyerPrice)), codec.Int32(offer.MinutesRemaining), codec.Int32(int32(offer.AveragePrice)), codec.Int32(offer.OfferCount))
}
