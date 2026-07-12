package refresh

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeValidatesEmptyPayload verifies MESSENGER_FRIENDS validation.
func TestDecodeValidatesEmptyPayload(t *testing.T) {
	if err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatal(err)
	}
}
