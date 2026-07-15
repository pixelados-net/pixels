package cfhsanctionstatus

import "testing"

// TestEncodeUsesHeader verifies compatibility header.
func TestEncodeUsesHeader(t *testing.T) {
	packet, err := Encode()
	if err != nil || packet.Header != Header {
		t.Fatal(err)
	}
}
