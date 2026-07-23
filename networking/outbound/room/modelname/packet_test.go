package modelname

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies ROOM_MODEL_NAME packet construction.
func TestEncode(t *testing.T) {
	packet, err := Encode("model_a", 7)
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		t.Fatalf("decode packet: %v", err)
	}
	if packet.Header != Header || values[1].Int32 != 7 {
		t.Fatalf("unexpected packet %#v values %#v", packet, values)
	}
}
