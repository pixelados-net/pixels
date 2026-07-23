package list

import "testing"

// TestEncode verifies USER_IGNORED framing.
func TestEncode(t *testing.T) {
	packet, err := Encode([]string{"demo", "alice"})
	if err != nil || packet.Header != Header || len(packet.Payload) == 0 {
		t.Fatalf("encode packet=%#v err=%v", packet, err)
	}
}
