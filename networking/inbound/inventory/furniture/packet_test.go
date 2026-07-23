package furniture

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies REQUEST_FURNITURE_INVENTORY decoding.
func TestDecode(t *testing.T) {
	if _, err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatalf("decode packet: %v", err)
	}
}

// TestDecodeRejectsUnexpectedPayload verifies trailing bytes are rejected.
func TestDecodeRejectsUnexpectedPayload(t *testing.T) {
	_, err := Decode(codec.Packet{Header: Header, Payload: []byte{1}})
	if !errors.Is(err, codec.ErrUnexpectedPayload) {
		t.Fatalf("expected unexpected payload, got %v", err)
	}
}

// TestDecodeRejectsUnexpectedHeader verifies header validation.
func TestDecodeRejectsUnexpectedHeader(t *testing.T) {
	_, err := Decode(codec.Packet{Header: Header + 1})
	if !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("expected unexpected header, got %v", err)
	}
}
