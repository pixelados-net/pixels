package stop

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies UNIT_TYPING_STOP validation.
func TestDecode(t *testing.T) {
	if err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatalf("decode typing stop: %v", err)
	}
	if err := Decode(codec.Packet{Header: Header + 1}); err == nil {
		t.Fatal("expected header rejection")
	}
}
