package error

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestEncode verifies exact room-ad error encoding.
func TestEncode(t *testing.T) {
	p, e := Encode(2, "bad")
	if e != nil {
		t.Fatal(e)
	}
	v, e := codec.DecodePacketExact(p, Definition)
	if e != nil || p.Header != Header || v[0].Int32 != 2 || v[1].String != "bad" {
		t.Fatalf("values=%#v err=%v", v, e)
	}
}
