package own

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestDecode verifies own-listings request header validation.
func TestDecode(t *testing.T) {
	if err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatal(err)
	}
	if err := Decode(codec.Packet{}); err != codec.ErrUnexpectedHeader {
		t.Fatalf("unexpected %v", err)
	}
}
