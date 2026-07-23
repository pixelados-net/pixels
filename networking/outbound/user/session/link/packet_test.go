package link

import "testing"

// TestEncode verifies in-client link encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode("navigator/goto/7")
	if err != nil || packet.Header != Header {
		t.Fatalf("encode link: %#v, %v", packet, err)
	}
}
