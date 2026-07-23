package failure

import "testing"

// TestEncode verifies PET_SCRATCH_FAILED output.
func TestEncode(t *testing.T) {
	packet, err := Encode(1, 3)
	if err != nil || packet.Header != Header || len(packet.Payload) != 8 {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
