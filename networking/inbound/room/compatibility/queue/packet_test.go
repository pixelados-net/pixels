package queue

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestDecode verifies the exact unused queue shape.
func TestDecode(t *testing.T) {
	p, e := codec.NewPacket(Header, Definition, codec.Int32(2))
	if e != nil {
		t.Fatal(e)
	}
	v, e := Decode(p)
	if e != nil || v != 2 {
		t.Fatalf("value=%d err=%v", v, e)
	}
	if _, e = Decode(codec.Packet{Header: Header + 1, Payload: p.Payload}); e == nil {
		t.Fatal("expected header error")
	}
	if _, e = Decode(codec.Packet{Header: Header}); e == nil {
		t.Fatal("expected payload error")
	}
}
