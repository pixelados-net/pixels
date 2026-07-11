package toggle

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecodeRequiresEmptyPayload verifies mute-all packet strictness.
func TestDecodeRequiresEmptyPayload(t *testing.T) {
	if err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if err := Decode(codec.Packet{Header: Header, Payload: []byte{1}}); err == nil {
		t.Fatal("expected payload rejection")
	}
}
