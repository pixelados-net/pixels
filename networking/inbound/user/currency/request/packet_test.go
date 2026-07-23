package request

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeAcceptsEmptyRequest verifies the packet contract.
func TestDecodeAcceptsEmptyRequest(t *testing.T) {
	if _, err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatalf("decode: %v", err)
	}
}

// TestDecodeRejectsWrongHeaderAndPayload verifies exact packet validation.
func TestDecodeRejectsWrongHeaderAndPayload(t *testing.T) {
	if _, err := Decode(codec.Packet{Header: Header + 1}); !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("expected unexpected header, got %v", err)
	}
	if _, err := Decode(codec.Packet{Header: Header, Payload: []byte{1}}); !errors.Is(err, codec.ErrUnexpectedPayload) {
		t.Fatalf("expected unexpected payload, got %v", err)
	}
}
