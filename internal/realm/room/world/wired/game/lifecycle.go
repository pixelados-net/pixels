package game

import (
	"context"
	"sort"
	"time"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// schedule queues one event on the existing room-owned task lifecycle.
func (coordinator *Coordinator) schedule(active *roomlive.Room, event trigger.Event) {
	if coordinator.engine == nil {
		return
	}
	active.Schedule(0, func(now time.Time) {
		_, _ = coordinator.engine.Process(context.Background(), event, now)
	})
}

// scheduleOutcomes emits one team result per participant before game-ended.
func (coordinator *Coordinator) scheduleOutcomes(active *roomlive.Room, state State, winners map[int32]struct{}) {
	if len(winners) == 0 {
		coordinator.schedule(active, trigger.Event{Kind: trigger.GameEnded, RoomID: active.ID()})
		return
	}
	playerIDs := make([]int64, 0, len(state.Teams))
	for playerID := range state.Teams {
		playerIDs = append(playerIDs, playerID)
	}
	sort.Slice(playerIDs, func(left int, right int) bool { return playerIDs[left] < playerIDs[right] })
	for _, playerID := range playerIDs {
		team := state.Teams[playerID]
		kind := trigger.TeamLost
		if _, won := winners[team]; won {
			kind = trigger.TeamWon
		}
		coordinator.schedule(active, trigger.Event{Kind: kind, RoomID: active.ID(), ActorKind: trigger.ActorPlayer, ActorID: playerID, PlayerID: playerID, Team: team})
	}
	coordinator.schedule(active, trigger.Event{Kind: trigger.GameEnded, RoomID: active.ID()})
}

// winningTeams returns the sole non-empty winning team or none on a tie.
func winningTeams(state State) map[int32]struct{} {
	var best int64
	hasTeam := false
	for _, team := range state.Teams {
		if team < 1 || team > 4 {
			continue
		}
		if !hasTeam || state.TeamScores[team] > best {
			best = state.TeamScores[team]
			hasTeam = true
		}
	}
	result := make(map[int32]struct{})
	if !hasTeam {
		return result
	}
	for _, team := range state.Teams {
		if team < 1 || team > 4 {
			continue
		}
		if state.TeamScores[team] == best {
			result[team] = struct{}{}
		}
	}
	if len(result) != 1 {
		return map[int32]struct{}{}
	}
	return result
}

// resultsFor creates board-specific normalized submissions.
func resultsFor(mode record.HighscoreMode, state State, winners map[int32]struct{}) []record.HighscoreResult {
	if mode == record.HighscoreClassic {
		playerIDs := make([]int64, 0, len(state.Scores))
		for playerID := range state.Scores {
			playerIDs = append(playerIDs, playerID)
		}
		sort.Slice(playerIDs, func(left int, right int) bool { return playerIDs[left] < playerIDs[right] })
		results := make([]record.HighscoreResult, 0, len(playerIDs))
		for _, playerID := range playerIDs {
			score := state.Scores[playerID]
			_, won := winners[state.Teams[playerID]]
			results = append(results, record.HighscoreResult{PlayerIDs: []int64{playerID}, Score: score, Won: won})
		}
		return results
	}
	byTeam := make(map[int32][]int64)
	for playerID, team := range state.Teams {
		byTeam[team] = append(byTeam[team], playerID)
	}
	teams := make([]int, 0, len(byTeam))
	for team := range byTeam {
		teams = append(teams, int(team))
	}
	sort.Ints(teams)
	results := make([]record.HighscoreResult, 0, len(teams))
	for _, teamValue := range teams {
		team := int32(teamValue)
		playerIDs := byTeam[team]
		sort.Slice(playerIDs, func(left int, right int) bool { return playerIDs[left] < playerIDs[right] })
		_, won := winners[team]
		results = append(results, record.HighscoreResult{PlayerIDs: playerIDs, Score: state.TeamScores[team], Won: won})
	}
	return results
}
