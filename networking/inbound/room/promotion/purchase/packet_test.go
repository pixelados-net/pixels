package purchase

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestDecode verifies the exact room-ad purchase shape.
func TestDecode(t *testing.T) {
	p, e := codec.NewPacket(Header, Definition, codec.Int32(1), codec.Int32(2), codec.Int32(3), codec.String("QA"), codec.Bool(true), codec.String("Lab"), codec.Int32(4))
	if e != nil {
		t.Fatal(e)
	}
	v, e := Decode(p)
	if e != nil || v.PageID != 1 || v.OfferID != 2 || v.RoomID != 3 || v.Title != "QA" || !v.Extended || v.Description != "Lab" || v.CategoryID != 4 {
		t.Fatalf("value=%#v err=%v", v, e)
	}
	if _, e = Decode(codec.Packet{Header: Header + 1, Payload: p.Payload}); e == nil {
		t.Fatal("expected header error")
	}
	if _, e = Decode(codec.Packet{Header: Header}); e == nil {
		t.Fatal("expected payload error")
	}
}
