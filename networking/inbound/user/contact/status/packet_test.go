package status

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies the retired status packet remains strict.
func TestDecode(t *testing.T) {
	if _, err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatalf("decode status: %v", err)
	}
	if _, err := Decode(codec.Packet{Header: Header + 1}); !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("expected unexpected header, got %v", err)
	}
}
