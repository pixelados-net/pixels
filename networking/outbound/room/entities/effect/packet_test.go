package effect

import "testing"

// TestEncodeUsesProtocolHeader verifies the outbound packet identity.
func TestEncodeUsesProtocolHeader(t *testing.T) {
	packet, err := Encode(7, 101, 0)
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
