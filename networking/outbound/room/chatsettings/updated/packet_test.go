package updated

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeWritesChatSettings verifies chat settings encoding.
func TestEncodeWritesChatSettings(t *testing.T) {
	packet, _ := Encode(1, 2, 1, 50, 2)
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || values[3].Int32 != 50 {
		t.Fatalf("values=%#v err=%v", values, err)
	}
}
