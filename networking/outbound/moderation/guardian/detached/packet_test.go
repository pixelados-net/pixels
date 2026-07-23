package detached

import "testing"

// TestEncodeUsesHeader verifies the protocol identifier.
func TestEncodeUsesHeader(t *testing.T) {
	packet, err := Encode()
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
