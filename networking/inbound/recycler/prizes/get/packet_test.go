package get

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies compatibility prize requests remain safe.
func TestDecode(t *testing.T) {
	if err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatalf("decode prizes: %v", err)
	}
}
