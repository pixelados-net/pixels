package discount

import "testing"

// TestEncodeCreatesHeaderOnlyRuleset verifies the protocol acknowledgement.
func TestEncodeCreatesHeaderOnlyRuleset(t *testing.T) {
	packet, err := Encode()
	if err != nil || packet.Header != Header || len(packet.Payload) != 0 {
		t.Fatalf("packet=%#v error=%v", packet, err)
	}
}
