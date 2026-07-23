package searchresult

import (
	"testing"

	"github.com/niflaot/pixels/networking/outbound/navigator/browse/roomcard"
)

// BenchmarkEncode measures navigator search result encoding.
func BenchmarkEncode(b *testing.B) {
	rooms := make([]roomcard.Card, 0, 100)
	for index := range 100 {
		rooms = append(rooms, roomcard.Card{
			RoomID:       int32(index + 1),
			RoomName:     "Benchmark Room",
			OwnerID:      1,
			OwnerName:    "demo",
			UserCount:    int32(index % 25),
			MaxUserCount: 25,
			Description:  "Benchmark room payload.",
			Tags:         []string{"demo", "benchmark"},
			ShowOwner:    true,
		})
	}

	lists := []ResultList{{Code: "popular", Mode: 1, Rooms: rooms}}
	for b.Loop() {
		if _, err := Encode("hotel_view", "", lists); err != nil {
			b.Fatalf("encode packet: %v", err)
		}
	}
}
