package add

import (
	"testing"

	petdata "github.com/niflaot/pixels/networking/pet/data"
)

// TestEncode verifies USER_PET_ADD output.
func TestEncode(t *testing.T) {
	packet, err := Encode(petdata.Pet{ID: 1, Name: "Pixel", Figure: petdata.Figure{Color: "FFFFFF"}, Level: 1}, true)
	if err != nil || packet.Header != Header || len(packet.Payload) == 0 {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
