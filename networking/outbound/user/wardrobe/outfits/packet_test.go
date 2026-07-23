package outfits

import "testing"

// TestEncode verifies wardrobe list encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(1, []int32{2}, []string{"hd-180-1"}, []string{"M"})
	if err != nil || packet.Header != Header || len(packet.Payload) == 0 {
		t.Fatalf("encode outfits: %#v, %v", packet, err)
	}
	if _, err = Encode(1, []int32{2}, nil, nil); err == nil {
		t.Fatal("expected invalid parallel lists")
	}
}
