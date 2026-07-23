package entryinfo

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies ROOM_INFO_OWNER packet construction.
func TestEncode(t *testing.T) {
	packet, err := Encode(42, true)
	if err != nil {
		t.Fatalf("encode room entry info: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		t.Fatalf("decode room entry info: %v", err)
	}
	if packet.Header != Header || values[0].Int32 != 42 || !values[1].Boolean {
		t.Fatalf("unexpected room entry info packet=%#v values=%#v", packet, values)
	}
}
