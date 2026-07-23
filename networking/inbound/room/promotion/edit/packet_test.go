package edit

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestDecode verifies exact promotion edit decoding.
func TestDecode(t *testing.T) {
	p, e := codec.NewPacket(Header, Definition, codec.Int32(8), codec.String("New"), codec.String("Text"))
	if e != nil {
		t.Fatal(e)
	}
	v, e := Decode(p)
	if e != nil || v.EventID != 8 || v.Name != "New" || v.Description != "Text" {
		t.Fatalf("value=%#v err=%v", v, e)
	}
	if _, e = Decode(codec.Packet{Header: Header + 1, Payload: p.Payload}); e == nil {
		t.Fatal("expected header error")
	}
	if _, e = Decode(codec.Packet{Header: Header}); e == nil {
		t.Fatal("expected payload error")
	}
}
