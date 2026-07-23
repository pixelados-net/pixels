package success

import "testing"

// TestEncode verifies PET_NEST_BREEDING_SUCCESS output.
func TestEncode(t *testing.T) {
	packet, err := Encode(8, 2)
	if err != nil || packet.Header != Header || len(packet.Payload) != 8 {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
