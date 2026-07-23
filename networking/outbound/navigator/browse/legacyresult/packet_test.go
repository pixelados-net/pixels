package legacyresult

import (
	"testing"

	"github.com/niflaot/pixels/networking/codec"
	"github.com/niflaot/pixels/networking/outbound/navigator/browse/roomcard"
)

// TestEncode verifies the renderer-compatible prefix and empty-ad trailer.
func TestEncode(t *testing.T) {
	packet, err := Encode(1, "demo", []roomcard.Card{{RoomID: 7, RoomName: "Room", OwnerID: 1, OwnerName: "demo", MaxUserCount: 25}})
	if err != nil || packet.Header != Header {
		t.Fatalf("packet=%#v err=%v", packet, err)
	}
	values, remaining, err := codec.DecodePayload(nil, Definition, packet.Payload)
	if err != nil || values[0].Int32 != 1 || values[1].String != "demo" || values[2].Int32 != 1 || len(remaining) == 0 {
		t.Fatalf("values=%#v remaining=%d err=%v", values, len(remaining), err)
	}
}

// BenchmarkEncode50 measures bounded legacy room-card encoding.
func BenchmarkEncode50(b *testing.B) {
	rooms := make([]roomcard.Card, 50)
	for index := range rooms {
		rooms[index] = roomcard.Card{RoomID: int32(index + 1), RoomName: "Room", OwnerID: 1, OwnerName: "demo", MaxUserCount: 25}
	}
	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		_, _ = Encode(1, "", rooms)
	}
}
