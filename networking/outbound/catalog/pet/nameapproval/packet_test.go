package nameapproval

import "testing"

// TestEncode verifies name approval output.
func TestEncode(t *testing.T) {
	packet, err := Encode(0, "")
	if err != nil || packet.Header != Header || len(packet.Payload) != 6 {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
