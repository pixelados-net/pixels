package clicked

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestDecode verifies exact tab-click telemetry decoding.
func TestDecode(t *testing.T) {
	p, e := codec.NewPacket(Header, Definition, codec.Int32(3), codec.String("QA"), codec.Int32(2))
	if e != nil {
		t.Fatal(e)
	}
	v, e := Decode(p)
	if e != nil || v.RoomID != 3 || v.RoomName != "QA" || v.Source != 2 {
		t.Fatalf("value=%#v err=%v", v, e)
	}
	if _, e = Decode(codec.Packet{Header: Header + 1, Payload: p.Payload}); e == nil {
		t.Fatal("expected header error")
	}
	if _, e = Decode(codec.Packet{Header: Header}); e == nil {
		t.Fatal("expected payload error")
	}
}
