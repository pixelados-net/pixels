package proceed

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies strict retired NUX decoding.
func TestDecode(t *testing.T) {
	if _, err := Decode(codec.Packet{Header: Header}); err != nil {
		t.Fatalf("decode proceed: %v", err)
	}
	if _, err := Decode(codec.Packet{Header: Header, Payload: []byte{1}}); !errors.Is(err, codec.ErrUnexpectedPayload) {
		t.Fatalf("expected unexpected payload, got %v", err)
	}
}
