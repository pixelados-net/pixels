package add

import "testing"

// TestEncodeUsesProtocolHeader verifies the outbound packet identity.
func TestEncodeUsesProtocolHeader(t *testing.T) {
	packet, err := Encode(101, 0, 60, false)
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}

// TestDefinitionMatchesNitroParser verifies the incremental packet field count.
func TestDefinitionMatchesNitroParser(t *testing.T) {
	if len(Definition) != 4 {
		t.Fatalf("expected four fields, got %d", len(Definition))
	}
}
