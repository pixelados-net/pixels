package recommendedrooms

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the header-only request and strict validation.
func TestDecode(t *testing.T) {
	if err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatal(err)
	}
	if err := Decode(codec.Packet{Header: Header + 1}); err == nil {
		t.Fatal("expected unexpected-header error")
	}
	if err := Decode(codec.Packet{Header: Header, Payload: []byte{0}}); err == nil {
		t.Fatal("expected trailing-payload error")
	}
}
