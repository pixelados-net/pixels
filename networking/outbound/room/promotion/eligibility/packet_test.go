package eligibility

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestEncode verifies exact eligibility encoding.
func TestEncode(t *testing.T) {
	p, e := Encode(true, 0)
	if e != nil {
		t.Fatal(e)
	}
	v, e := codec.DecodePacketExact(p, Definition)
	if e != nil || p.Header != Header || !v[0].Boolean || v[1].Int32 != 0 {
		t.Fatalf("values=%#v err=%v", v, e)
	}
}
