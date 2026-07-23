package offers

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeWritesClubOfferCount verifies offer list framing.
func TestEncodeWritesClubOfferCount(t *testing.T) {
	packet, err := Encode([]Offer{{ID: 1, Name: "hc_31_days", PriceCredits: 25,
		PointsType: -1, Months: 1, DaysLeftAfterPurchase: 31, Year: 2026, Month: 8, Day: 12}})
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v error=%v", packet, err)
	}
	definition := codec.Definition{codec.Int32Field, codec.Int32Field, codec.StringField,
		codec.BooleanField, codec.Int32Field, codec.Int32Field, codec.Int32Field,
		codec.BooleanField, codec.Int32Field, codec.Int32Field, codec.BooleanField,
		codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field}
	values, remaining, err := codec.DecodePayload(nil, definition, packet.Payload)
	if err != nil || len(remaining) != 0 || values[0].Int32 != 1 || values[1].Int32 != 1 ||
		values[2].String != "hc_31_days" || values[4].Int32 != 25 || values[8].Int32 != 1 ||
		values[11].Int32 != 31 || values[12].Int32 != 2026 || values[13].Int32 != 8 || values[14].Int32 != 12 {
		t.Fatalf("values=%#v error=%v", values, err)
	}
}
