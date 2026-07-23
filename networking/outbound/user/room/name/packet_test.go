package name

import "testing"

// TestEncode verifies room-unit name encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(7, 3, "Pixel")
	if err != nil || packet.Header != Header {
		t.Fatalf("encode room name: %#v, %v", packet, err)
	}
}
