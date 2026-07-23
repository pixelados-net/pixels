package trynumber

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the exact header-only shape.
func TestDecode(t *testing.T) {
	if err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatal(err)
	}
}
