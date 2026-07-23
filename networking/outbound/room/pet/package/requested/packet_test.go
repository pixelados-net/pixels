package requested

import "testing"

// TestEncode verifies PET_OPEN_PACKAGE_REQUESTED output.
func TestEncode(t *testing.T) {
	packet, err := Encode(7, "0 0 FFFFFF 0")
	if err != nil || packet.Header != Header || len(packet.Payload) == 0 {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
