package permissions

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncodeWritesPermissionFields verifies Nitro's permission packet shape.
func TestEncodeWritesPermissionFields(t *testing.T) {
	packet, err := Encode(2, 100, true)
	if err != nil {
		t.Fatalf("encode permissions: %v", err)
	}
	values, err := codec.DecodePacketExact(packet, Definition)
	if err != nil {
		t.Fatalf("decode permissions: %v", err)
	}
	if values[0].Int32 != 2 || values[1].Int32 != 100 || !values[2].Boolean {
		t.Fatalf("unexpected permission values %#v", values)
	}
}

// BenchmarkEncode measures USER_PERMISSIONS packet encoding.
func BenchmarkEncode(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		packet, err := Encode(0, 100, true)
		if err != nil || packet.Header != Header {
			b.Fatalf("unexpected packet=%#v err=%v", packet, err)
		}
	}
}
