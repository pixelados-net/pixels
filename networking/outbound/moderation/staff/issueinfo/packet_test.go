package issueinfo

import "testing"

// TestEncodeUsesHeader verifies issue projection.
func TestEncodeUsesHeader(t *testing.T) {
	packet, err := Encode(Params{})
	if err != nil || packet.Header != Header {
		t.Fatal(err)
	}
}
