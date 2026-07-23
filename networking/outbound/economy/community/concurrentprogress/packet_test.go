package concurrentprogress

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the neutral concurrent users goal wire shape.
func TestEncode(t *testing.T) {
	packet, err := Encode(0, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || packet.Header != Header || len(packet.Payload) != 12 {
		t.Fatalf("packet=%+v err=%v", packet, err)
	}
	for _, value := range values {
		if value.Int32 != 0 {
			t.Fatalf("unexpected neutral values: %+v", values)
		}
	}
}
