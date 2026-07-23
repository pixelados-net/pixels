package training

import "testing"

// TestEncode verifies PET_TRAINING_PANEL output.
func TestEncode(t *testing.T) {
	packet, err := Encode(7, []int32{0, 1, 2}, []int32{0, 1})
	if err != nil || packet.Header != Header || len(packet.Payload) == 0 {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
