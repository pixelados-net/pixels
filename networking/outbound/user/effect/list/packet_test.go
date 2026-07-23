package list

import "testing"

// TestEncodeUsesProtocolHeader verifies the outbound packet identity.
func TestEncodeUsesProtocolHeader(t *testing.T) {
	packet, err := Encode([]Effect{{Type: 101, Duration: 60, InactiveEffectsInInventory: 1}})
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
