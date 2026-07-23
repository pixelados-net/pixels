package result

import "testing"

// TestEncode verifies USER_IGNORED_RESULT framing.
func TestEncode(t *testing.T) {
	packet, err := Encode(Ignored, "demo")
	if err != nil || packet.Header != Header || len(packet.Payload) == 0 {
		t.Fatalf("encode packet=%#v err=%v", packet, err)
	}
}
