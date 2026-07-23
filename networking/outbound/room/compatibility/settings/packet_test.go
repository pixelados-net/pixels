package settings

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestEncode verifies the superseded error shape.
func TestEncode(t *testing.T) {
	p, e := Encode(4, 2)
	if e != nil {
		t.Fatal(e)
	}
	v, e := codec.DecodePacketExact(p, Definition)
	if e != nil || p.Header != Header || v[0].Int32 != 4 || v[1].Int32 != 2 {
		t.Fatalf("values=%#v err=%v", v, e)
	}
}
