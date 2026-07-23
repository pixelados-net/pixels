package basic

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeReadsSelectedOffer verifies the extension offer id is required.
func TestDecodeReadsSelectedOffer(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(12))
	if err != nil {
		t.Fatalf("encode fixture: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.OfferID != 12 {
		t.Fatalf("payload=%#v error=%v", payload, err)
	}
}
