package wallitems

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestEncode verifies complete wall snapshots include owners and item records.
func TestEncode(t *testing.T) {
	packet, err := Encode([]Owner{{ID: 7, Name: "demo"}}, []Item{{ID: 42, SpriteID: 9, WallPosition: ":w=2,3 l=4,5 r", ExtraData: "0", OwnerID: 7}})
	if err != nil || packet.Header != Header {
		t.Fatalf("encode wall snapshot: %#v err=%v", packet, err)
	}
	count, rest, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field}, packet.Payload)
	if err != nil || count[0].Int32 != 1 {
		t.Fatalf("unexpected owner count %#v err=%v", count, err)
	}
	owner, rest, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field, codec.StringField}, rest)
	if err != nil || owner[1].String != "demo" {
		t.Fatalf("unexpected owner %#v err=%v", owner, err)
	}
	itemCount, rest, err := codec.DecodePayload(nil, codec.Definition{codec.Int32Field}, rest)
	if err != nil || itemCount[0].Int32 != 1 {
		t.Fatalf("unexpected item count %#v err=%v", itemCount, err)
	}
	item, rest, err := codec.DecodePayload(nil, codec.Definition{codec.StringField, codec.Int32Field, codec.StringField, codec.StringField, codec.Int32Field, codec.Int32Field, codec.Int32Field}, rest)
	if err != nil || item[0].String != "42" || item[2].String != ":w=2,3 l=4,5 r" || len(rest) != 0 {
		t.Fatalf("unexpected item %#v rest=%d err=%v", item, len(rest), err)
	}
}
