package roominfo

import "testing"

// TestEncodeUsesHeader verifies room information projection.
func TestEncodeUsesHeader(t *testing.T) {
	packet, err := Encode(Params{})
	if err != nil || packet.Header != Header {
		t.Fatal(err)
	}
}
