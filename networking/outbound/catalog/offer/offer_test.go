package offer

import (
	"strings"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestAppendPageEncodesLimitedProduct verifies Nitro page offer shape.
func TestAppendPageEncodesLimitedProduct(t *testing.T) {
	payload, err := AppendPage(nil, fixtureOffer())
	if err != nil {
		t.Fatalf("append page offer: %v", err)
	}
	if len(payload) == 0 || payload[len(payload)-3] != 0 || payload[len(payload)-2] != 0 || payload[len(payload)-1] != 0 {
		t.Fatalf("unexpected page trailer %v", payload)
	}
}

// TestAppendOfferRejectsOversizedStrings verifies serializer failures.
func TestAppendOfferRejectsOversizedStrings(t *testing.T) {
	oversized := strings.Repeat("x", 1<<16)
	if _, err := AppendPage(nil, Offer{LocalizationID: oversized}); err == nil {
		t.Fatal("expected oversized localization error")
	}
	value := fixtureOffer()
	value.Products[0].ExtraData = oversized
	if _, err := AppendPurchase(nil, value); err == nil {
		t.Fatal("expected oversized product data error")
	}
	value = Offer{PreviewImage: oversized}
	if _, err := AppendPage(nil, value); err == nil {
		t.Fatal("expected oversized preview error")
	}
}

// TestAppendRegularProductOmitsLimitedFields verifies regular product shape.
func TestAppendRegularProductOmitsLimitedFields(t *testing.T) {
	value := fixtureOffer()
	value.Products[0].Limited = false
	if payload, err := AppendPage(nil, value); err != nil || len(payload) == 0 {
		t.Fatalf("unexpected payload size=%d error %v", len(payload), err)
	}
}

// TestAppendPurchaseOmitsPageTrailer verifies Nitro purchase offer shape.
func TestAppendPurchaseOmitsPageTrailer(t *testing.T) {
	payload, err := AppendPurchase(nil, fixtureOffer())
	if err != nil {
		t.Fatalf("append purchase offer: %v", err)
	}
	values, _, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field, codec.StringField}, payload)
	if err != nil || values[0].Int32 != 7 || values[1].String != "catalog.item.chair" {
		t.Fatalf("unexpected payload values %#v error %v", values, err)
	}
}

// fixtureOffer creates one limited test offer.
func fixtureOffer() Offer {
	return Offer{ID: 7, LocalizationID: "catalog.item.chair", CostCredits: 2, Giftable: true,
		Products: []Product{{Type: "s", ClassID: 1, ExtraData: "0", Amount: 1, Limited: true, LimitedStack: 10, LimitedRemaining: 9}}}
}
