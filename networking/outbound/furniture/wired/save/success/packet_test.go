package success

import "testing"

// TestEncode verifies the empty WIRED_SAVE packet.
func TestEncode(t *testing.T) {
	packet, err := Encode()
	if err != nil || packet.Header != Header || len(packet.Payload) != 0 {
		t.Fatalf("unexpected success %#v %v", packet, err)
	}
}
