package staffpick

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestDecode verifies strict staff-pick decoding.
func TestDecode(t *testing.T) {
	p, e := codec.NewPacket(Header, Definition, codec.Int32(7))
	if e != nil {
		t.Fatal(e)
	}
	v, e := Decode(p)
	if e != nil || v != 7 {
		t.Fatalf("value=%d err=%v", v, e)
	}
	if _, e = Decode(codec.Packet{Header: Header + 1, Payload: p.Payload}); e == nil {
		t.Fatal("expected header error")
	}
	if _, e = Decode(codec.Packet{Header: Header}); e == nil {
		t.Fatal("expected payload error")
	}
}
