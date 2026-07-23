package figure

import (
	"testing"

	petdata "github.com/niflaot/pixels/networking/pet/data"
)

// TestEncode verifies PET_FIGURE_UPDATE output.
func TestEncode(t *testing.T) {
	packet, err := Encode(2, 1, petdata.Figure{Color: "FFFFFF"}, true, false)
	if err != nil || packet.Header != Header || len(packet.Payload) == 0 {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
}
