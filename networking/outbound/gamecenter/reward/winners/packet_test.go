package winners

import "testing"

// TestEncode verifies the variable winners response.
func TestEncode(t *testing.T) {
	packet, err := Encode(2, []Winner{{Name: "demo", Figure: "hd-180-1", Gender: "M", Rank: 1, Score: 99}})
	if err != nil || packet.Header != Header {
		t.Fatalf("unexpected packet: %+v %v", packet, err)
	}
}
