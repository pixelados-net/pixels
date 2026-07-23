package halloffame

import "testing"

// TestEncode verifies the goal code and empty entry count.
func TestEncode(t *testing.T) {
	packet, err := Encode("habboFameComp")
	if err != nil || packet.Header != Header || len(packet.Payload) != 19 {
		t.Fatalf("unexpected packet=%+v err=%v", packet, err)
	}
}
