package muted

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeWritesRemainingSeconds verifies mute encoding.
func TestEncodeWritesRemainingSeconds(t *testing.T) {
	packet, err := Encode(300)
	if err != nil {
		t.Fatal(err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || values[0].Int32 != 300 {
		t.Fatalf("unexpected values %#v err=%v", values, err)
	}
}
