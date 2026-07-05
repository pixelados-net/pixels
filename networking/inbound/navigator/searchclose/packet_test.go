package searchclose

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies NAVIGATOR_SEARCH_CLOSE decoding.
func TestDecode(t *testing.T) {
	packet, err := codec.NewPacket(Header, Definition, codec.String("popular"))
	if err != nil {
		t.Fatalf("new packet: %v", err)
	}
	payload, err := Decode(packet)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if payload.Code != "popular" {
		t.Fatalf("unexpected payload %#v", payload)
	}
}
