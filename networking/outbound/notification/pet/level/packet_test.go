package level

import (
	"testing"

	petdata "github.com/niflaot/pixels/networking/pet/data"
)

// TestEncode verifies PET_LEVEL_NOTIFICATION output.
func TestEncode(t *testing.T) {
	packet, err := Encode(1, "Pixel", 2, petdata.Figure{Color: "FFFFFF"})
	if err != nil || packet.Header != Header || len(packet.Payload) == 0 {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
