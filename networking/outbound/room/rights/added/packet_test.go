package added

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeWritesAddedRight verifies rights-list add encoding.
func TestEncodeWritesAddedRight(t *testing.T) {
	packet, err := Encode(9, 7, "Alice")
	if err != nil {
		t.Fatal(err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || values[0].Int32 != 9 || values[1].Int32 != 7 || values[2].String != "Alice" {
		t.Fatalf("unexpected values %#v err=%v", values, err)
	}
}
