package breeds

import "testing"

// TestEncode verifies palette output.
func TestEncode(t *testing.T) {
	packet, err := Encode("pet0", []Palette{{TypeID: 0, BreedID: 1, PaletteID: 2, Sellable: true}})
	if err != nil || packet.Header != Header || len(packet.Payload) == 0 {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
