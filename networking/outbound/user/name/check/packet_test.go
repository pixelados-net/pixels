package check

import "testing"

// TestEncode verifies name-check result encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(0, "Pixel", nil)
	if err != nil || packet.Header != Header || len(packet.Payload) == 0 {
		t.Fatalf("encode name check: %#v, %v", packet, err)
	}
}
