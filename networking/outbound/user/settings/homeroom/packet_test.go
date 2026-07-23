package homeroom

import "testing"

// TestEncode verifies home-room projection.
func TestEncode(t *testing.T) {
	packet, err := Encode(7, 0)
	if err != nil || packet.Header != Header {
		t.Fatalf("encode home room: %#v, %v", packet, err)
	}
}
