package issuedeleted

import "testing"

// TestEncodeUsesHeader verifies the protocol identifier.
func TestEncodeUsesHeader(t *testing.T) {
	packet, err := Encode("x")
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
