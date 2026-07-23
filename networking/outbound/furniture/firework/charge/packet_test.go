package charge

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestEncode verifies the documented provisional shape.
func TestEncode(t *testing.T) {
	p, e := Encode(9, true)
	if e != nil {
		t.Fatal(e)
	}
	v, e := codec.DecodePacketExact(p, Definition)
	if e != nil || p.Header != Header || v[0].Int32 != 9 || !v[1].Boolean {
		t.Fatalf("values=%#v err=%v", v, e)
	}
}
