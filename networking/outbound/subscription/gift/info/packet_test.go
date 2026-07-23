package info

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
	catalogoffer "github.com/niflaot/pixels/networking/outbound/catalog/offer"
)

// TestEncodeWritesOffersAndGiftMetadata verifies Nitro's two club-gift sections.
func TestEncodeWritesOffersAndGiftMetadata(t *testing.T) {
	packet, err := Encode(2, 1, []Gift{{Offer: catalogoffer.Offer{ID: 9, LocalizationID: "monthly_sofa"},
		DaysRequired: 30, Selectable: true}})
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v error=%v", packet, err)
	}
	values, remaining, err := codec.DecodePayload(nil, codec.Definition{
		codec.Int32Field, codec.Int32Field, codec.Int32Field,
		codec.Int32Field, codec.StringField, codec.BooleanField, codec.Int32Field,
		codec.Int32Field, codec.Int32Field, codec.BooleanField, codec.Int32Field,
		codec.Int32Field, codec.BooleanField, codec.BooleanField, codec.StringField,
		codec.Int32Field, codec.Int32Field, codec.BooleanField, codec.Int32Field,
		codec.BooleanField,
	}, packet.Payload)
	if err != nil || len(remaining) != 0 || values[0].Int32 != 2 || values[1].Int32 != 1 ||
		values[2].Int32 != 1 || values[3].Int32 != 9 || values[4].String != "monthly_sofa" ||
		values[15].Int32 != 1 || values[16].Int32 != 9 || values[18].Int32 != 30 || !values[19].Boolean {
		t.Fatalf("values=%#v error=%v", values, err)
	}
}
