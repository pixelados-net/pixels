package init

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies NAVIGATOR_INIT decoding.
func TestDecode(t *testing.T) {
	if _, err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatalf("decode packet: %v", err)
	}
}

// TestDecodeRejectsUnexpectedHeader verifies packet header validation.
func TestDecodeRejectsUnexpectedHeader(t *testing.T) {
	_, err := Decode(codec.Packet{Header: Header + 1})
	if !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("expected unexpected header error, got %v", err)
	}
}
