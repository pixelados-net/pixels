package bots

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode validates empty USER_BOTS requests and strict headers.
func TestDecode(t *testing.T) {
	if err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if err := Decode(codec.Packet{Header: Header + 1}); !errors.Is(err, codec.ErrUnexpectedHeader) {
		t.Fatalf("unexpected header error=%v", err)
	}
	if err := Decode(codec.Packet{Header: Header, Payload: []byte{1}}); err == nil {
		t.Fatal("expected payload rejection")
	}
}
