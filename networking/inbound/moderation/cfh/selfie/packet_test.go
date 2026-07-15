package selfie

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestDecodeRejectsWrongHeader verifies header validation.
func TestDecodeRejectsWrongHeader(t *testing.T) {
	packet := codec.Packet{Header: Header + 1}
	if _, err := Decode(packet); err != codec.ErrUnexpectedHeader {
		t.Fatalf("err=%v", err)
	}
}
