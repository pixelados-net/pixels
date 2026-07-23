package clothing

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestDecode verifies the exact retired packet shape.
func TestDecode(t *testing.T) {
	p, e := codec.NewPacket(Header, Definition, codec.Int32(7), codec.String("M"), codec.String("hd-1"))
	if e != nil {
		t.Fatal(e)
	}
	v, e := Decode(p)
	if e != nil || v.ObjectID != 7 || v.Gender != "M" || v.Look != "hd-1" {
		t.Fatalf("value=%#v err=%v", v, e)
	}
	if _, e = Decode(codec.Packet{Header: Header + 1, Payload: p.Payload}); e == nil {
		t.Fatal("expected header error")
	}
	if _, e = Decode(codec.Packet{Header: Header}); e == nil {
		t.Fatal("expected payload error")
	}
}
