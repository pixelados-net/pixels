package game

import (
	"context"
	"math"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// ExecuteGame executes team and score effects.
func (service *Service) ExecuteGame(ctx context.Context, operation effect.GameOperation, node *configuration.Node, event trigger.Event) (effect.Result, error) {
	if event.PlayerID <= 0 {
		return effect.Result{Status: effect.Skipped}, nil
	}
	service.mutex.Lock()
	defer service.mutex.Unlock()
	state := service.stateLocked(event.RoomID)
	switch operation {
	case effect.JoinTeam:
		team := first(node.Parameters.Values)
		state.Teams[event.PlayerID] = team
		service.projectTeam(event.RoomID, event.PlayerID, team)
	case effect.LeaveTeam:
		delete(state.Teams, event.PlayerID)
		service.projectTeam(event.RoomID, event.PlayerID, 0)
	case effect.GiveScore:
		return service.giveWiredScore(event, node, state), nil
	case effect.GiveTeamScore:
		return service.giveTeamScore(event, node, state), nil
	case effect.ResetHighscore:
		return service.resetHighscore(ctx, event, node)
	default:
		return effect.Result{Status: effect.Blocked}, nil
	}
	return effect.Result{Status: effect.Applied}, nil
}

// giveWiredScore applies a bounded individual score effect once per configured use.
func (service *Service) giveWiredScore(event trigger.Event, node *configuration.Node, state *State) effect.Result {
	if !state.Running {
		return effect.Result{Status: effect.Skipped}
	}
	amount, limit := int64(first(node.Parameters.Values)), second(node.Parameters.Values)
	if limit < 1 {
		limit = 1
	}
	uses := state.WiredUses[event.PlayerID]
	if uses == nil {
		uses = make(map[int64]int32)
		state.WiredUses[event.PlayerID] = uses
	}
	if uses[node.ItemID] >= limit {
		return effect.Result{Status: effect.Skipped}
	}
	uses[node.ItemID]++
	previous := state.Scores[event.PlayerID]
	state.Scores[event.PlayerID] = saturatedAdd(previous, amount)
	state.WiredScores[event.PlayerID] = saturatedAdd(state.WiredScores[event.PlayerID], amount)
	return scoreResult(event, previous, state.Scores[event.PlayerID])
}

// giveTeamScore applies a bounded team score to an active match.
func (service *Service) giveTeamScore(_ trigger.Event, node *configuration.Node, state *State) effect.Result {
	if !state.Running {
		return effect.Result{Status: effect.Skipped}
	}
	team := first(node.Parameters.Values)
	if team < 1 || team > 4 {
		return effect.Result{Status: effect.Blocked}
	}
	state.TeamScores[team] = saturatedAdd(state.TeamScores[team], int64(second(node.Parameters.Values)))
	return effect.Result{Status: effect.Applied}
}

// resetHighscore deletes only configured boards from the active room.
func (service *Service) resetHighscore(ctx context.Context, event trigger.Event, node *configuration.Node) (effect.Result, error) {
	if service.highscores == nil || len(node.Targets) == 0 {
		return effect.Result{Status: effect.Blocked}, nil
	}
	ids := make([]int64, 0, len(node.Targets))
	for _, target := range node.Targets {
		ids = append(ids, target.ItemID)
	}
	service.mutex.Unlock()
	_, err := service.highscores.Reset(ctx, event.RoomID, ids)
	service.mutex.Lock()
	if err != nil {
		return effect.Result{Status: effect.Blocked}, err
	}
	return effect.Result{Status: effect.Applied}, nil
}

// scoreResult creates a threshold event for trigger-side crossing checks.
func scoreResult(event trigger.Event, previous int64, current int64, _ ...*configuration.Node) effect.Result {
	derived := event
	derived.ID, derived.Kind = 0, trigger.ScoreAchieved
	derived.PreviousScore, derived.Score = previous, current
	return effect.Result{Status: effect.Applied, Derived: []trigger.Event{derived}}
}

// AddScore atomically changes one running participant score.
func (service *Service) AddScore(roomID int64, playerID int64, amount int64) (int64, int64, bool) {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	state := service.rooms[roomID]
	if state == nil || !state.Running {
		return 0, 0, false
	}
	team, participating := state.Teams[playerID]
	if !participating {
		return 0, 0, false
	}
	previous := state.Scores[playerID]
	current := saturatedAdd(previous, amount)
	state.Scores[playerID], state.TeamScores[team] = current, saturatedAdd(state.TeamScores[team], amount)
	return previous, current, true
}

// saturatedAdd prevents score overflow from wrapping ranking order.
func saturatedAdd(value int64, delta int64) int64 {
	if delta > 0 && value > math.MaxInt64-delta {
		return math.MaxInt64
	}
	if delta < 0 && value < math.MinInt64-delta {
		return math.MinInt64
	}
	return value + delta
}

// first returns the first setting or zero.
func first(values []int32) int32 {
	if len(values) == 0 {
		return 0
	}
	return values[0]
}

// second returns the second setting or zero.
func second(values []int32) int32 {
	if len(values) < 2 {
		return 0
	}
	return values[1]
}
