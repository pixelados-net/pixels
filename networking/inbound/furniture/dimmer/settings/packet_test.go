package settings

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the empty settings request and header validation.
func TestDecode(t *testing.T) {
	if _, err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatalf("decode settings: %v", err)
	}
	if _, err := Decode(codec.Packet{Header: Header + 1}); err == nil {
		t.Fatal("expected header error")
	}
}
