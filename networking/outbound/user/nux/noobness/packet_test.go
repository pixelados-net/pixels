package noobness

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies old-identity projection.
func TestEncode(t *testing.T) {
	packet, err := Encode(OldIdentity)
	if err != nil {
		t.Fatalf("encode noobness: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil || values[0].Int32 != OldIdentity {
		t.Fatalf("decode noobness: %#v, %v", values, err)
	}
}
