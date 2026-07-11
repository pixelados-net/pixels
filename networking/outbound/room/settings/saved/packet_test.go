package saved

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeWritesRoomID verifies save confirmation encoding.
func TestEncodeWritesRoomID(t *testing.T) {
	packet, _ := Encode(8)
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || values[0].Int32 != 8 {
		t.Fatalf("values=%#v err=%v", values, err)
	}
}
