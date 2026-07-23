package received

import (
	"testing"

	petdata "github.com/niflaot/pixels/networking/pet/data"
)

// TestEncode verifies PET_RECEIVED output.
func TestEncode(t *testing.T) {
	packet, err := Encode(false, petdata.Pet{ID: 2, Name: "Pixel", Figure: petdata.Figure{Color: "FFFFFF"}, Level: 1})
	if err != nil || packet.Header != Header || len(packet.Payload) == 0 {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
