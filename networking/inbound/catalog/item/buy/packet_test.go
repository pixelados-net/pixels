package buy

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeVerifiesNitroComposerOrder verifies PURCHASE_FROM_CATALOG decoding.
func TestDecodeVerifiesNitroComposerOrder(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.Int32(3), codec.Int32(8), codec.String("color=1"), codec.Int32(2))
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil || payload.PageID != 3 || payload.OfferID != 8 || payload.ExtraData != "color=1" || payload.Amount != 2 {
		t.Fatalf("unexpected payload %#v error %v", payload, err)
	}
}

// TestDecodeRejectsHeader verifies packet identity validation.
func TestDecodeRejectsHeader(t *testing.T) {
	if _, err := Decode(codec.Packet{}); err != codec.ErrUnexpectedHeader {
		t.Fatalf("expected unexpected header, got %v", err)
	}
}
