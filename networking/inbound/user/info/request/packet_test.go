package request

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the USER_INFO request.
func TestDecode(t *testing.T) {
	if _, err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatalf("decode request: %v", err)
	}
}
