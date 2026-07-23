package search

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestDecode verifies exact room-ad search decoding.
func TestDecode(t *testing.T) {
	p, e := codec.NewPacket(Header, Definition, codec.Int32(4), codec.Int32(20))
	if e != nil {
		t.Fatal(e)
	}
	v, e := Decode(p)
	if e != nil || v.CategoryID != 4 || v.Offset != 20 {
		t.Fatalf("value=%#v err=%v", v, e)
	}
	if _, e = Decode(codec.Packet{Header: Header + 1, Payload: p.Payload}); e == nil {
		t.Fatal("expected header error")
	}
	if _, e = Decode(codec.Packet{Header: Header}); e == nil {
		t.Fatal("expected payload error")
	}
}
