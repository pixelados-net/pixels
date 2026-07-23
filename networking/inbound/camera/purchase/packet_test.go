package purchase

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies compatibility with empty and extra payloads.
func TestDecode(t *testing.T) {
	for _, payload := range [][]byte{nil, {0, 1, 'x'}} {
		if err := Decode(codec.Packet{Header: Header, Payload: payload}); err != nil {
			t.Fatal(err)
		}
	}
}
