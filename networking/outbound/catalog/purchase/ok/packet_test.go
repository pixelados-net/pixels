package ok

import (
	"strings"
	"testing"

	"github.com/niflaot/pixels/networking/outbound/catalog/offer"
)

// TestEncode verifies PURCHASE_OK packet construction.
func TestEncode(t *testing.T) {
	packet, err := Encode(offer.Offer{ID: 8})
	if err != nil || packet.Header != Header || len(packet.Payload) == 0 {
		t.Fatalf("unexpected packet %#v error %v", packet, err)
	}
}

// TestEncodeReportsInvalidOffer verifies purchase serializer failures.
func TestEncodeReportsInvalidOffer(t *testing.T) {
	if _, err := Encode(offer.Offer{LocalizationID: strings.Repeat("x", 1<<16)}); err == nil {
		t.Fatal("expected oversized localization error")
	}
}
