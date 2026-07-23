package vote

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the exact retired packet shape.
func TestDecode(t *testing.T) {
	if err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatal(err)
	}
	if err := Decode(codec.Packet{Header: Header + 1}); err == nil {
		t.Fatal("expected header error")
	}
}
