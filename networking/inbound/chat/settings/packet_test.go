package settings

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies GET_SOUND_SETTINGS header and payload validation.
func TestDecode(t *testing.T) {
	if err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if err := Decode(codec.Packet{Header: Header + 1}); err == nil {
		t.Fatal("expected header rejection")
	}
}
