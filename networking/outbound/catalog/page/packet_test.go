package page

import (
	"strings"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/outbound/catalog/offer"
)

// TestEncodeVerifiesPagePrefix verifies CATALOG_PAGE layout and localization fields.
func TestEncodeVerifiesPagePrefix(t *testing.T) {
	packet, err := Encode(4, "NORMAL", "default_3x3", Localization{Images: []string{"a"}, Texts: []string{"Chairs"}}, []offer.Offer{}, -1)
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	values, _, err := codec.DecodePayload(nil, codec.Definition{
		codec.Int32Field, codec.StringField, codec.StringField, codec.Int32Field, codec.StringField, codec.Int32Field, codec.StringField,
	}, packet.Payload)
	if err != nil || packet.Header != Header || values[0].Int32 != 4 || values[2].String != "default_3x3" || values[6].String != "Chairs" {
		t.Fatalf("unexpected values %#v error %v", values, err)
	}
}

// TestEncodeIncludesOffersAndReportsInvalidStrings verifies dynamic page sections.
func TestEncodeIncludesOffersAndReportsInvalidStrings(t *testing.T) {
	value := offer.Offer{ID: 2, LocalizationID: "catalog.item.chair"}
	packet, err := Encode(4, "NORMAL", "default_3x3", Localization{}, []offer.Offer{value}, -1)
	if err != nil || len(packet.Payload) == 0 {
		t.Fatalf("unexpected packet %#v error %v", packet, err)
	}
	oversized := strings.Repeat("x", 1<<16)
	if _, err := Encode(4, "NORMAL", "default_3x3", Localization{Images: []string{oversized}}, nil, -1); err == nil {
		t.Fatal("expected oversized image error")
	}
	value.LocalizationID = oversized
	if _, err := Encode(4, "NORMAL", "default_3x3", Localization{}, []offer.Offer{value}, -1); err == nil {
		t.Fatal("expected oversized offer error")
	}
}
