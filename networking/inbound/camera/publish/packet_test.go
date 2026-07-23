package publish

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the header-only request.
func TestDecode(t *testing.T) {
	if err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatal(err)
	}
}
