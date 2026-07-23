package outside

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies strict header-only alias validation.
func TestDecode(t *testing.T) {
	if err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatal(err)
	}
	if err := Decode(codec.Packet{Header: Header + 1}); err == nil {
		t.Fatal("expected unexpected header")
	}
	if err := Decode(codec.Packet{Header: Header, Payload: []byte{0}}); err == nil {
		t.Fatal("expected trailing payload")
	}
}
