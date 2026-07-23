package requests

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeValidatesEmptyPayload verifies GET_FRIEND_REQUESTS validation.
func TestDecodeValidatesEmptyPayload(t *testing.T) {
	if err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatal(err)
	}
}
