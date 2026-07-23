package level

import "testing"

// TestEncode verifies PET_LEVEL_UPDATE output.
func TestEncode(t *testing.T) {
	packet, err := Encode(2, 1, 3)
	if err != nil || packet.Header != Header || len(packet.Payload) != 12 {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
