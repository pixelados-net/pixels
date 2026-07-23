package eventcats

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies GET_USER_EVENT_CATS decoding.
func TestDecode(t *testing.T) {
	if _, err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatalf("decode packet: %v", err)
	}
}

// TestDecodeRejectsWrongHeader verifies header validation.
func TestDecodeRejectsWrongHeader(t *testing.T) {
	if _, err := Decode(codec.Packet{Header: Header + 1}); !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("expected unexpected header, got %v", err)
	}
}
