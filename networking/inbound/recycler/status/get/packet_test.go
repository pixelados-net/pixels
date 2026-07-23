package get

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the empty recycler status request.
func TestDecode(t *testing.T) {
	if err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatalf("decode status: %v", err)
	}
	if err := Decode(codec.Packet{Header: Header, Payload: []byte{1}}); err == nil {
		t.Fatal("expected trailing payload rejection")
	}
}
