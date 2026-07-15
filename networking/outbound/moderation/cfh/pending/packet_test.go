package pending

import "testing"

// TestEncodeUsesHeader verifies pending projection.
func TestEncodeUsesHeader(t *testing.T) {
	packet, err := Encode(nil)
	if err != nil || packet.Header != Header {
		t.Fatal(err)
	}
}
