package forward

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies FORWARD_TO_SOME_ROOM decoding.
func TestDecode(t *testing.T) {
	if _, err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatalf("decode packet: %v", err)
	}
}
