package nosuchflat

import (
	"testing"
)

// TestEncode verifies NO_SUCH_FLAT packet construction.
func TestEncode(t *testing.T) {
	packet, err := Encode(0)
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	if packet.Header != Header {
		t.Fatalf("expected header %d, got %d", Header, packet.Header)
	}
}
