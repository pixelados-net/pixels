package tool

import "testing"

// TestEncodeUsesHeader verifies tool initialization.
func TestEncodeUsesHeader(t *testing.T) {
	packet, err := Encode(nil, nil, nil, Permissions{})
	if err != nil || packet.Header != Header {
		t.Fatal(err)
	}
}
