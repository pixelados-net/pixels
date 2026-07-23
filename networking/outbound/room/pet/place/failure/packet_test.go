package failure

import "testing"

// TestEncode verifies PET_PLACING_ERROR output.
func TestEncode(t *testing.T) {
	packet, err := Encode(2)
	if err != nil || packet.Header != Header || len(packet.Payload) != 4 {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
