package userinfo

import "testing"

// TestEncodeUsesHeader verifies user information projection.
func TestEncodeUsesHeader(t *testing.T) {
	packet, err := Encode(Params{})
	if err != nil || packet.Header != Header {
		t.Fatal(err)
	}
}
