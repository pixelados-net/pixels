package reportmessage

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestDecodeValidatesHeaderAndPayload verifies strict empty decoding.
func TestDecodeValidatesHeaderAndPayload(t *testing.T) {
	if err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatal(err)
	}
	if err := Decode(codec.Packet{Header: Header + 1}); err != codec.ErrUnexpectedHeader {
		t.Fatalf("unexpected error %v", err)
	}
}
