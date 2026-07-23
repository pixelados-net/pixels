package record

import (
	marketcore "github.com/niflaot/pixels/internal/realm/marketplace/core"
	marketrecord "github.com/niflaot/pixels/internal/realm/marketplace/record"
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestAppendOffer verifies shared listing encoding.
func TestAppendOffer(t *testing.T) {
	offer := marketcore.Offer{Listing: marketrecord.Listing{ID: 1, RawPrice: 10}, BuyerPrice: 11, AveragePrice: 12, OfferCount: 3}
	payload, err := AppendOffer(nil, offer, false)
	definition := codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}
	values, decodeErr := codec.DecodePacketExact(codec.Packet{Payload: payload}, definition)
	if err != nil || decodeErr != nil || values[8].Int32 != 12 || values[9].Int32 != 3 {
		t.Fatalf("values=%#v err=%v decode=%v", values, err, decodeErr)
	}
	payload, err = AppendOffer(nil, offer, true)
	ownDefinition := codec.Definition{codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field}
	values, decodeErr = codec.DecodePacketExact(codec.Packet{Payload: payload}, ownDefinition)
	if err != nil || decodeErr != nil || values[6].Int32 != 10 {
		t.Fatalf("own values=%#v err=%v decode=%v", values, err, decodeErr)
	}
}
