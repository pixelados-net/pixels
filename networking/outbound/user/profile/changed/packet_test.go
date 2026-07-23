package changed

import "testing"

// TestEncode verifies profile-change encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(7)
	if err != nil || packet.Header != Header {
		t.Fatalf("encode changed: %#v, %v", packet, err)
	}
}
