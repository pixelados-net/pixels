package respect

import "testing"

// TestEncode verifies respect projection encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(7, 12)
	if err != nil || packet.Header != Header {
		t.Fatalf("encode respect: %#v, %v", packet, err)
	}
}
