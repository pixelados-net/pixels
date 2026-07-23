package event

import (
	"github.com/niflaot/pixels/networking/codec"
	"testing"
)

// TestEncode verifies renderer-order room event encoding.
func TestEncode(t *testing.T) {
	in := Data{AdID: 1, OwnerAvatarID: 2, OwnerAvatarName: "demo", RoomID: 3, EventType: 4, Name: "QA", Description: "Lab", MinutesSinceCreation: 5, MinutesUntilExpiration: 115, CategoryID: 6}
	p, e := Encode(in)
	if e != nil {
		t.Fatal(e)
	}
	v, e := codec.DecodePacketExact(p, Definition)
	if e != nil || p.Header != Header || v[0].Int32 != 1 || v[2].String != "demo" || v[9].Int32 != 6 {
		t.Fatalf("values=%#v err=%v", v, e)
	}
}
