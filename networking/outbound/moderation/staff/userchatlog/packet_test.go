package userchatlog

import "testing"

// TestEncodeUsesHeader verifies grouped chatlog projection.
func TestEncodeUsesHeader(t *testing.T) {
	packet, err := Encode(1, "a", nil)
	if err != nil || packet.Header != Header {
		t.Fatal(err)
	}
}
