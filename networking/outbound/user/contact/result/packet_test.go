package result

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies CHANGE_EMAIL_RESULT wire compatibility.
func TestEncode(t *testing.T) {
	packet, err := Encode(1)
	if err != nil {
		t.Fatalf("encode result: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || values[0].Int32 != 1 {
		t.Fatalf("decode result: %#v, %v", values, err)
	}
}
