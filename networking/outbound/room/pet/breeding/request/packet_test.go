package request

import "testing"

// TestEncode verifies the variable breeding confirmation request.
func TestEncode(t *testing.T) {
	packet, err := Encode(9, Parent{ID: 1, Name: "A", Figure: "0 0 FFFFFF 0", OwnerName: "demo"}, Parent{ID: 2, Name: "B", Figure: "0 0 FFFFFF 0", OwnerName: "alice"}, []RarityCategory{{Chance: 100, Breeds: []int32{0, 1}}}, 0)
	if err != nil || packet.Header != Header || len(packet.Payload) == 0 {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
