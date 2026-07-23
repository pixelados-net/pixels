package dicevalue

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies FURNITURE_STATE_2 encoding.
func TestEncode(t *testing.T) {
	packet, err := Encode(17, 6)
	if err != nil {
		t.Fatalf("encode packet: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || values[0].Int32 != 17 || values[1].Int32 != 6 {
		t.Fatalf("unexpected values %#v err=%v", values, err)
	}
}
