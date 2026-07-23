package banned

import "testing"

// TestEncode verifies the banned packet header.
func TestEncode(t *testing.T) {
	packet, err := Encode("banned")
	if err != nil || packet.Header != Header {
		t.Fatalf("encode banned: %#v, %v", packet, err)
	}
}
