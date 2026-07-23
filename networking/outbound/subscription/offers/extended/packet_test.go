package extended

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
	cluboffers "github.com/niflaot/pixels/networking/outbound/subscription/offers"
)

// TestEncodeAppendsExtensionPricing verifies Nitro's four extended fields.
func TestEncodeAppendsExtensionPricing(t *testing.T) {
	packet, err := Encode(cluboffers.Offer{ID: 2, Name: "hc_90_days", Months: 2}, 80, 10, 5, 20)
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v error=%v", packet, err)
	}
	definition := codec.Definition{
		codec.Int32Field, codec.StringField, codec.BooleanField, codec.Int32Field,
		codec.Int32Field, codec.Int32Field, codec.BooleanField, codec.Int32Field,
		codec.Int32Field, codec.BooleanField, codec.Int32Field, codec.Int32Field,
		codec.Int32Field, codec.Int32Field, codec.Int32Field, codec.Int32Field,
		codec.Int32Field, codec.Int32Field,
	}
	values, remaining, err := codec.DecodePayload(nil, definition, packet.Payload)
	if err != nil || len(remaining) != 0 || values[14].Int32 != 80 || values[15].Int32 != 10 ||
		values[16].Int32 != 5 || values[17].Int32 != 20 {
		t.Fatalf("values=%#v remaining=%d error=%v", values, len(remaining), err)
	}
}
