package unaccept

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestDecode verifies the header-only packet.
func TestDecode(t *testing.T) {
	if err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatal(err)
	}
}
