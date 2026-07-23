package categorycounts

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies GET_CATEGORIES_WITH_USER_COUNT decoding.
func TestDecode(t *testing.T) {
	if _, err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatalf("decode packet: %v", err)
	}
}
