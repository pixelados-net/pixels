package welcomeresult

import "testing"

// TestEncode verifies the deprecated welcome result header.
func TestEncode(t *testing.T) {
	packet, err := Encode(1)
	if err != nil || packet.Header != Header {
		t.Fatalf("encode welcome result: %#v, %v", packet, err)
	}
}
