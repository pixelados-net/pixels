package closeissue

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestDecodeRejectsWrongHeader verifies strict headers.
func TestDecodeRejectsWrongHeader(t *testing.T) {
	if _, err := Decode(codec.Packet{}); err != codec.ErrUnexpectedHeader {
		t.Fatal(err)
	}
}
