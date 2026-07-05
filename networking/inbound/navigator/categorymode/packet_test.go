package categorymode

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies NAVIGATOR_CATEGORY_LIST_MODE decoding.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.String("popular"), codec.Int32(1))
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if payload.Category != "popular" || payload.ListMode != 1 {
		t.Fatalf("unexpected payload %#v", payload)
	}
}
