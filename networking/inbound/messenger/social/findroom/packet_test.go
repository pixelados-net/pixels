package findroom

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeValidatesEmptyPayload verifies FIND_NEW_FRIENDS validation.
func TestDecodeValidatesEmptyPayload(t *testing.T) {
	if err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatal(err)
	}
}
