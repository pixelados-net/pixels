package weekly

import "testing"

// TestEncodeEmpty verifies a valid product-free reward response.
func TestEncodeEmpty(t *testing.T) {
	packet, err := EncodeEmpty(4, 100, false)
	if err != nil || packet.Header != Header || len(packet.Payload) != 13 {
		t.Fatalf("unexpected packet: %+v %v", packet, err)
	}
}
