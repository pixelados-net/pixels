package incomplete

import "testing"

// TestEncode verifies the retired NUX header-only packet.
func TestEncode(t *testing.T) {
	packet, err := Encode()
	if err != nil || packet.Header != Header || len(packet.Payload) != 0 {
		t.Fatalf("encode incomplete: %#v, %v", packet, err)
	}
}
