package updates

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeValidatesEmptyPayload verifies FRIEND_LIST_UPDATE validation.
func TestDecodeValidatesEmptyPayload(t *testing.T) {
	if err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatal(err)
	}
}
