package game

import (
	"context"
	"math"
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// TestServiceLifecycleAndExclusiveTeams verifies lifecycle, exclusive team membership, and scoring.
func TestServiceLifecycleAndExclusiveTeams(t *testing.T) {
	service := New()
	event := trigger.Event{RoomID: 7, PlayerID: 42}
	join := &configuration.Node{Parameters: configuration.Parameters{Values: []int32{2}}}
	if result, err := service.ExecuteGame(context.Background(), effect.JoinTeam, join, event); err != nil || result.Status != effect.Applied {
		t.Fatalf("pre-game join failed: result=%+v err=%v", result, err)
	}
	if !service.Start(7) || service.Start(7) {
		t.Fatal("expected exactly one successful start")
	}
	if team, found := service.Team(7, 42); !found || team != 2 {
		t.Fatalf("start dropped preselected team: team=%d found=%v", team, found)
	}
	join.Parameters.Values[0] = 4
	_, _ = service.ExecuteGame(context.Background(), effect.JoinTeam, join, event)
	if team, found := service.Team(7, 42); !found || team != 4 {
		t.Fatalf("exclusive team replacement failed: team=%d found=%v", team, found)
	}
	previous, current, ok := service.AddScore(7, 42, 9)
	if !ok || previous != 0 || current != 9 {
		t.Fatalf("score failed: previous=%d current=%d ok=%v", previous, current, ok)
	}
	state, _ := service.Snapshot(7)
	if state.TeamScores[4] != 9 {
		t.Fatalf("team score=%d, want 9", state.TeamScores[4])
	}
	if !service.End(7) || service.End(7) {
		t.Fatal("expected exactly one successful end")
	}
	if _, _, ok = service.AddScore(7, 42, 1); ok {
		t.Fatal("ended game accepted score")
	}
	service.Reset(7)
	if _, found := service.Snapshot(7); found {
		t.Fatal("reset retained room state")
	}
}

// TestWinningTeamsRejectsTies verifies team outcomes are not order-dependent.
func TestWinningTeamsRejectsTies(t *testing.T) {
	state := State{Teams: map[int64]int32{1: 1, 2: 2}}
	state.TeamScores[1], state.TeamScores[2] = 10, 10
	if winners := winningTeams(state); len(winners) != 0 {
		t.Fatalf("tie winners=%v", winners)
	}
	state.TeamScores[2] = 9
	winners := winningTeams(state)
	if _, found := winners[1]; len(winners) != 1 || !found {
		t.Fatalf("sole winners=%v", winners)
	}
}

// TestScoreThresholdCrossesOnceAndSaturates verifies threshold and overflow policy.
func TestScoreThresholdCrossesOnceAndSaturates(t *testing.T) {
	event := trigger.Event{RoomID: 1, PlayerID: 2}
	node := &configuration.Node{Parameters: configuration.Parameters{Values: []int32{6, 10}}}
	firstResult := scoreResult(event, 5, 11, node)
	secondResult := scoreResult(event, 11, 17, node)
	if len(firstResult.Derived) != 1 || len(secondResult.Derived) != 1 {
		t.Fatalf("unexpected threshold events: first=%d second=%d", len(firstResult.Derived), len(secondResult.Derived))
	}
	if got := saturatedAdd(math.MaxInt64-1, 4); got != math.MaxInt64 {
		t.Fatalf("positive overflow=%d", got)
	}
	if got := saturatedAdd(math.MinInt64+1, -4); got != math.MinInt64 {
		t.Fatalf("negative overflow=%d", got)
	}
}

// TestRemovePlayerClearsStaleParticipation verifies room leave cannot leak a team into reconnection.
func TestRemovePlayerClearsStaleParticipation(t *testing.T) {
	service := New()
	if !service.JoinTeam(9, 7, 3) || !service.Start(9) {
		t.Fatal("prepare participant")
	}
	if _, _, changed := service.AddScore(9, 7, 5); !changed {
		t.Fatal("score participant")
	}
	if !service.RemovePlayer(9, 7) {
		t.Fatal("remove participant")
	}
	if _, found := service.Team(9, 7); found {
		t.Fatal("stale team remained")
	}
	snapshot, _ := service.Snapshot(9)
	if _, found := snapshot.Scores[7]; found {
		t.Fatal("stale score remained")
	}
}
