package model

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies ROOM_MODEL decoding.
func TestDecode(t *testing.T) {
	if _, err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatalf("decode packet: %v", err)
	}
}
