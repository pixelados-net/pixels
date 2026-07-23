package state

import "testing"

// TestEncode verifies PET_BREEDING output.
func TestEncode(t *testing.T) {
	packet, err := Encode(3, 1, 2)
	if err != nil || packet.Header != Header || len(packet.Payload) != 12 {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
