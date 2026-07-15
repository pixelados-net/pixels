package sanctionstatus

import "testing"

// TestEncodeUsesHeader verifies sanction projection.
func TestEncodeUsesHeader(t *testing.T) {
	packet, err := Encode(Params{Name: "ALERT"})
	if err != nil || packet.Header != Header {
		t.Fatal(err)
	}
}
