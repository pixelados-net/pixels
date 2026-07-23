package failure

import "testing"

// TestEncode verifies PET_GO_TO_BREEDING_NEST_FAILURE output.
func TestEncode(t *testing.T) {
	packet, err := Encode(6)
	if err != nil || packet.Header != Header || len(packet.Payload) != 4 {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
