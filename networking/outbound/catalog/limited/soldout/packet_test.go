package soldout

import "testing"

// TestEncode verifies LIMITED_SOLD_OUT packet construction.
func TestEncode(t *testing.T) {
	packet, err := Encode()
	if err != nil || packet.Header != Header || len(packet.Payload) != 0 {
		t.Fatalf("unexpected packet %#v error %v", packet, err)
	}
}
