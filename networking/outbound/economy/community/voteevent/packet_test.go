package voteevent

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the neutral vote acknowledgement wire shape.
func TestEncode(t *testing.T) {
	packet, err := Encode(false)
	if err != nil {
		t.Fatal(err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || packet.Header != Header || len(packet.Payload) != 1 || values[0].Boolean {
		t.Fatalf("packet=%+v values=%+v err=%v", packet, values, err)
	}
}
