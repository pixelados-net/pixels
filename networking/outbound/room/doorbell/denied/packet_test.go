package denied

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies ROOM_DOORBELL_REJECTED encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode("Demo")
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || values[0].String != "Demo" {
		t.Fatalf("unexpected values %#v err=%v", values, err)
	}
}
