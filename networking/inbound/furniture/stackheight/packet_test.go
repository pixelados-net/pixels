package stackheight

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestDecode verifies centimeter precision and the automatic-height sentinel.
func TestDecode(t *testing.T) {
	for _, height := range []int32{123, AutoHeight} {
		packet, err := codec.NewPacket(Header, Definition, codec.Int32(7), codec.Int32(height))
		if err != nil {
			t.Fatal(err)
		}
		payload, err := Decode(packet)
		if err != nil || payload.ItemID != 7 || payload.Height != height {
			t.Fatalf("payload=%#v err=%v", payload, err)
		}
	}
}
