package selected

import "testing"

// TestEncodeUsesProtocolHeader verifies the outbound packet identity.
func TestEncodeUsesProtocolHeader(t *testing.T) {
	packet, err := Encode(101)
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
