package level

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeWritesRightsLevel verifies room rights encoding.
func TestEncodeWritesRightsLevel(t *testing.T) {
	packet, err := Encode(Rights)
	if err != nil {
		t.Fatal(err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || values[0].Int32 != Rights {
		t.Fatalf("unexpected values %#v err=%v", values, err)
	}
	if Owner != 4 || Moderator != 5 {
		t.Fatalf("unexpected Nitro controller levels owner=%d moderator=%d", Owner, Moderator)
	}
}
