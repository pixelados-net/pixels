package removed

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeWritesRemovedRight verifies rights-list removal encoding.
func TestEncodeWritesRemovedRight(t *testing.T) {
	packet, err := Encode(9, 7)
	if err != nil {
		t.Fatal(err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || values[0].Int32 != 9 || values[1].Int32 != 7 {
		t.Fatalf("unexpected values %#v err=%v", values, err)
	}
}
