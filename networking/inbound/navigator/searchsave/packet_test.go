package searchsave

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies NAVIGATOR_SEARCH_SAVE decoding.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.String("hotel_view"), codec.String("demo"))
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if payload.Code != "hotel_view" || payload.Data != "demo" {
		t.Fatalf("unexpected payload %#v", payload)
	}
}
