package category

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestEncode verifies the unused category shape.
func TestEncode(t *testing.T) {
	p, e := Encode(3)
	if e != nil {
		t.Fatal(e)
	}
	v, e := codec.DecodePacketExact(p, Definition)
	if e != nil || p.Header != Header || v[0].Int32 != 3 {
		t.Fatalf("values=%#v err=%v", v, e)
	}
}
