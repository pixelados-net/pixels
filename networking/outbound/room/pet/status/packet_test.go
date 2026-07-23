package status

import "testing"

// TestEncode verifies PET_STATUS output.
func TestEncode(t *testing.T) {
	packet, err := Encode(2, 1, true, false, false, true)
	if err != nil || packet.Header != Header || len(packet.Payload) != 12 {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
