package remove

import "testing"

// TestEncode verifies USER_PET_REMOVE output.
func TestEncode(t *testing.T) {
	packet, err := Encode(4)
	if err != nil || packet.Header != Header || len(packet.Payload) != 4 {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
