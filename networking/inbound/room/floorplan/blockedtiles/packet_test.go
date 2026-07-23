package blockedtiles

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeValidatesEmptyRequest verifies occupied tile request framing.
func TestDecodeValidatesEmptyRequest(t *testing.T) {
	if _, err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatalf("decode request: %v", err)
	}
	if _, err := Decode(codec.Packet{Header: Header + 1}); err != codec.ErrUnexpectedHeader {
		t.Fatalf("expected header error, got %v", err)
	}
}
