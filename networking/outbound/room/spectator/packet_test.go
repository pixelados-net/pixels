package spectator

import "testing"

// TestEncode verifies the header-only spectator response.
func TestEncode(t *testing.T) {
	packet, err := Encode()
	if err != nil || packet.Header != Header || len(packet.Payload) != 0 {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
