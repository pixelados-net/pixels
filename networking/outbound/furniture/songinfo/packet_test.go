package songinfo

import "testing"

// TestEncode verifies the empty song count.
func TestEncode(t *testing.T) {
	packet, err := Encode()
	if err != nil || packet.Header != Header || len(packet.Payload) != 4 {
		t.Fatalf("unexpected packet=%+v err=%v", packet, err)
	}
}
