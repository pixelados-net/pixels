package activated

import "testing"

// TestEncodeUsesProtocolHeader verifies the outbound packet identity.
func TestEncodeUsesProtocolHeader(t *testing.T) {
	packet, err := Encode(101, 60, false)
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}

// TestDefinitionMatchesNitroParser verifies the activation packet field count.
func TestDefinitionMatchesNitroParser(t *testing.T) {
	if len(Definition) != 3 {
		t.Fatalf("expected three fields, got %d", len(Definition))
	}
}
