package promoted

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestDecode verifies the exact retired packet shape.
func TestDecode(t *testing.T) {
	p, e := codec.NewPacket(Header, Definition, codec.String("official"))
	if e != nil {
		t.Fatal(e)
	}
	v, e := Decode(p)
	if e != nil || v != "official" {
		t.Fatalf("value=%q err=%v", v, e)
	}
	if _, e = Decode(codec.Packet{Header: Header + 1, Payload: p.Payload}); e == nil {
		t.Fatal("expected header error")
	}
	if _, e = Decode(codec.Packet{Header: Header}); e == nil {
		t.Fatal("expected payload error")
	}
}
