package purchase

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestEncode verifies room-ad purchase list encoding.
func TestEncode(t *testing.T) {
	p, e := Encode(true, []Room{{ID: 7, Name: "QA", Promoted: true}})
	if e != nil {
		t.Fatal(e)
	}
	v, rest, e := codec.DecodePacket(p, Definition)
	if e != nil || p.Header != Header || !v[0].Boolean || v[1].Int32 != 1 {
		t.Fatalf("values=%#v err=%v", v, e)
	}
	r, rest, e := codec.DecodePayload(nil, RoomDefinition, rest)
	if e != nil || len(rest) != 0 || r[0].Int32 != 7 || r[1].String != "QA" || !r[2].Boolean {
		t.Fatalf("room=%#v rest=%v err=%v", r, rest, e)
	}
}
