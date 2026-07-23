package wiring

import (
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// TestActorKindMapsEveryRoomEntity verifies movement adapters preserve unit kinds.
func TestActorKindMapsEveryRoomEntity(t *testing.T) {
	cases := []struct {
		kind worldunit.Kind
		want trigger.ActorKind
	}{
		{kind: worldunit.KindPlayer, want: trigger.ActorPlayer},
		{kind: worldunit.KindBot, want: trigger.ActorBot},
		{kind: worldunit.KindPet, want: trigger.ActorPet},
	}
	for _, test := range cases {
		if actual := actorKind(test.kind); actual != test.want {
			t.Fatalf("kind %d mapped to %d, want %d", test.kind, actual, test.want)
		}
	}
}

// TestNearAvatarMatchesClassicReachBoundary verifies crossing distance below two tiles.
func TestNearAvatarMatchesClassicReachBoundary(t *testing.T) {
	origin := grid.MustPoint(4, 4)
	if !nearAvatar(origin, grid.MustPoint(5, 5)) {
		t.Fatal("expected diagonal neighbor to be reached")
	}
	if nearAvatar(origin, grid.MustPoint(6, 4)) {
		t.Fatal("expected two-tile distance not to be reached")
	}
}
