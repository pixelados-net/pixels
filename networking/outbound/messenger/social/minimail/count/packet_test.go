package count

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies the deprecated MiniMail count wire.
func TestEncode(t *testing.T) {
	packet, err := Encode(3)
	if err != nil {
		t.Fatalf("encode count: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || values[0].Int32 != 3 {
		t.Fatalf("decode count: %#v, %v", values, err)
	}
}
