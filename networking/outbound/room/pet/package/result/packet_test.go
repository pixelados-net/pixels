package result

import "testing"

// TestEncode verifies PET_OPEN_PACKAGE_RESULT output.
func TestEncode(t *testing.T) {
	packet, err := Encode(7, 0, "")
	if err != nil || packet.Header != Header || len(packet.Payload) != 10 {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
