package entrytile

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies GET_ROOM_ENTRY_TILE decoding.
func TestDecode(t *testing.T) {
	if _, err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatalf("decode packet: %v", err)
	}
}

// TestDecodeRejectsWrongHeader verifies header validation.
func TestDecodeRejectsWrongHeader(t *testing.T) {
	_, err := Decode(codec.Packet{Header: 1})
	if !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("expected unexpected header, got %v", err)
	}
}

// TestDecodeRejectsInvalidPayload verifies exact payload validation.
func TestDecodeRejectsInvalidPayload(t *testing.T) {
	_, err := Decode(codec.Packet{Header: Header, Payload: []byte{1}})
	if err == nil {
		t.Fatal("expected payload error")
	}
}
