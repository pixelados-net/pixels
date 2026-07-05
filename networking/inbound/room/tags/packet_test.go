package tags

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies room tag packet decoding.
func TestDecode(t *testing.T) {
	payload, err := Decode(codec.Packet{Header: PopularHeader})
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if payload.Header != PopularHeader {
		t.Fatalf("unexpected payload %#v", payload)
	}
}
