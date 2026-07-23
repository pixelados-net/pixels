package confirmresult

import "testing"

// TestEncode verifies PET_CONFIRM_BREEDING_RESULT output.
func TestEncode(t *testing.T) {
	packet, err := Encode(7, 0)
	if err != nil || packet.Header != Header || len(packet.Payload) != 8 {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
