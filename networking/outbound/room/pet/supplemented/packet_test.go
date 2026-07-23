package supplemented

import "testing"

// TestEncode verifies PET_SUPPLEMENT output.
func TestEncode(t *testing.T) {
	packet, err := Encode(1, 2, 3)
	if err != nil || packet.Header != Header || len(packet.Payload) != 12 {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
