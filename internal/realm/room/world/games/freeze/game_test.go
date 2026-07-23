package freeze

import (
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
)

// TestExplosion verifies orthogonal, diagonal, massive, and blocked rays.
func TestExplosion(t *testing.T) {
	center := grid.MustPoint(5, 5)
	if got := len(Explosion(center, 2, false, false, nil)); got != 9 {
		t.Fatalf("orthogonal count %d", got)
	}
	if got := len(Explosion(center, 2, true, false, nil)); got != 9 {
		t.Fatalf("diagonal count %d", got)
	}
	if got := len(Explosion(center, 2, false, true, nil)); got != 17 {
		t.Fatalf("massive count %d", got)
	}
	valid := func(point grid.Point) bool { return point.X != 6 }
	if got := len(Explosion(center, 2, false, false, valid)); got != 7 {
		t.Fatalf("blocked count %d", got)
	}
}

// TestPlayerPowerUpsAndHit verifies caps, shielding, freeze, and loss floors.
func TestPlayerPowerUpsAndHit(t *testing.T) {
	now := time.Unix(100, 0)
	player := Player{Lives: 2, Snowballs: 2, Radius: 2}
	player.ApplyPowerUp(LifeUp, 5, 3, now, 10*time.Second, true)
	player.ApplyPowerUp(Shield, 5, 3, now, 10*time.Second, true)
	player.ApplyPowerUp(Shield, 5, 3, now.Add(time.Second), 10*time.Second, true)
	if player.Lives != 3 || !player.ProtectedUntil.Equal(now.Add(20*time.Second)) {
		t.Fatalf("unexpected power-up state: %+v", player)
	}
	if player.Hit(now.Add(5*time.Second), 5*time.Second, 5, 3) {
		t.Fatal("shielded player was hit")
	}
	if !player.Hit(now.Add(21*time.Second), 5*time.Second, 5, 3) || player.Lives != 2 || player.Snowballs != 1 || player.Radius != 1 {
		t.Fatalf("unexpected hit: %+v", player)
	}
}

// TestEveryPowerUp verifies every configured Freeze reward mutation.
func TestEveryPowerUp(t *testing.T) {
	now := time.Unix(100, 0)
	tests := []struct {
		// power identifies the applied reward.
		power PowerUp
		// verify checks its unique mutation.
		verify func(Player) bool
	}{
		{RangeUp, func(player Player) bool { return player.Radius == 2 }},
		{BallUp, func(player Player) bool { return player.Snowballs == 2 }},
		{Diagonal, func(player Player) bool { return player.Diagonal }},
		{Massive, func(player Player) bool { return player.Massive }},
		{LifeUp, func(player Player) bool { return player.Lives == 2 }},
		{Shield, func(player Player) bool { return player.ProtectedUntil.Equal(now.Add(10 * time.Second)) }},
	}
	for _, test := range tests {
		player := Player{Lives: 1, Snowballs: 1, Radius: 1}
		player.ApplyPowerUp(test.power, 5, 3, now, 10*time.Second, false)
		if !test.verify(player) {
			t.Errorf("power=%d player=%+v", test.power, player)
		}
	}
}

// TestFreezePointsPenalizesFriendlyFire verifies team-aware scoring.
func TestFreezePointsPenalizesFriendlyFire(t *testing.T) {
	if FreezePoints(1, 2, 10) != 10 || FreezePoints(1, 1, 10) != -10 || FreezePoints(0, 1, 10) != 10 {
		t.Fatal("unexpected friendly-fire scoring")
	}
}

// TestNativeVisualStates verifies values consumed by FurnitureIceStormLogic.
func TestNativeVisualStates(t *testing.T) {
	center := grid.MustPoint(5, 5)
	if ArmedState(1) != 2000 || ResetState(0) != 11 || ResetState(2) != 211 {
		t.Fatal("unexpected armed or reset state")
	}
	if DropState(Shield) != 7000 || CollectedState(Shield) != 17000 || Distance(center, grid.MustPoint(7, 4)) != 2 {
		t.Fatal("unexpected power-up state or distance")
	}
}

// TestDropHonorsChanceAndSpreadsContiguousIdentifiers verifies stable reward selection.
func TestDropHonorsChanceAndSpreadsContiguousIdentifiers(t *testing.T) {
	if _, found := Drop(10, 1, 0); found {
		t.Fatal("zero chance produced a reward")
	}
	seenDrop, seenEmpty := false, false
	for blockID := int64(961300); blockID < 961332; blockID++ {
		power, found := Drop(blockID, 1, 33)
		seenDrop, seenEmpty = seenDrop || found, seenEmpty || !found
		if found && (power < RangeUp || power > Shield) {
			t.Fatalf("invalid power %d", power)
		}
	}
	if !seenDrop || !seenEmpty {
		t.Fatalf("contiguous ids drop=%v empty=%v", seenDrop, seenEmpty)
	}
	if _, found := Drop(10, 1, 100); !found {
		t.Fatal("full chance omitted a reward")
	}
}

// TestApproachPointsStartsWithNearestNeighbor verifies block-click routing.
func TestApproachPointsStartsWithNearestNeighbor(t *testing.T) {
	points := ApproachPoints(grid.MustPoint(5, 5), grid.MustPoint(5, 8))
	if len(points) != 8 || Distance(grid.MustPoint(5, 5), points[0]) != 1 || Distance(grid.MustPoint(5, 8), points[0]) != 2 {
		t.Fatalf("unexpected approach points: %v", points)
	}
}

// TestWinningTeams verifies last-team-standing and score ties.
func TestWinningTeams(t *testing.T) {
	winners := WinningTeams([]Player{{Team: 1, Lives: 1}, {Team: 2, Lives: 0, Score: 99}})
	if len(winners) != 1 || winners[0] != 1 {
		t.Fatalf("unexpected survivor: %v", winners)
	}
	winners = WinningTeams([]Player{{Team: 1, Score: 5}, {Team: 2, Score: 5}})
	if len(winners) != 2 {
		t.Fatalf("expected tie: %v", winners)
	}
}

// BenchmarkFreezeTick measures explosion geometry without scheduler allocations.
func BenchmarkFreezeTick(b *testing.B) {
	center := grid.MustPoint(20, 20)
	valid := func(point grid.Point) bool { return point.X < 40 && point.Y < 40 }
	b.ReportAllocs()
	for range b.N {
		_ = Explosion(center, 5, false, true, valid)
	}
}
