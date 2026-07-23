package updated

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeWritesVisualizationSettings verifies thickness encoding.
func TestEncodeWritesVisualizationSettings(t *testing.T) {
	packet, _ := Encode(true, 1, -1)
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || !values[0].Boolean || values[2].Int32 != -1 {
		t.Fatalf("values=%#v err=%v", values, err)
	}
}
