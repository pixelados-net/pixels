package attached

import "testing"

// TestEncodeUsesHeader verifies the protocol identifier.
func TestEncodeUsesHeader(t *testing.T) {
	packet, err := Encode(true, 1, "x", 1)
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
