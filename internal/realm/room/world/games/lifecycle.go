package games

import (
	"context"
	"strconv"
	"strings"
	"time"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/games/freeze"
	wiredgame "github.com/niflaot/pixels/internal/realm/room/world/wired/game"
)

// Score records one durable player result from a completed room game.
type Score struct {
	// ID identifies one persisted player result.
	ID int64
	// RoomID identifies the room where the game ran.
	RoomID int64
	// Kind identifies banzai, freeze, football, tag, or wired.
	Kind string
	// StartedAt stores the authoritative game start instant.
	StartedAt time.Time
	// PlayerID identifies the participant.
	PlayerID int64
	// Team stores the participant team color.
	Team int32
	// Score stores the non-Wired achievement score.
	Score int64
	// TeamScore stores the final aggregate team score.
	TeamScore int64
}

// ScorePage stores a bounded history query.
type ScorePage struct {
	// Entries stores the page in descending persistence order.
	Entries []Score
	// NextID stores the cursor for the following page, or zero at the end.
	NextID int64
}

// ScoreStore persists and reads room game history.
type ScoreStore interface {
	// Save inserts one completed match atomically.
	Save(context.Context, []Score) error
	// List returns a cursor-paginated room history.
	List(context.Context, int64, int64, int) (ScorePage, error)
}

// finish ends, persists, and publishes one completed match.
func (service *Service) finish(ctx context.Context, active *roomlive.Room, startedAt time.Time) error {
	service.mutex.Lock()
	state := service.states[active.ID()]
	if state == nil || state.finishing {
		service.mutex.Unlock()
		return nil
	}
	state.finishing = true
	if state.timer != nil {
		state.timer.Reset()
	}
	service.mutex.Unlock()
	snapshot, found := service.wired.Snapshot(active.ID())
	if !found {
		return wiredgame.ErrGameUnavailable
	}
	if err := service.coordinator.End(ctx, active.ID()); err != nil {
		service.mutex.Lock()
		state.finishing = false
		service.mutex.Unlock()
		return err
	}
	kind := roomGameKind(active)
	if kind == "freeze" {
		service.cleanupFreezeMatch(ctx, active)
	}
	service.metrics.ended[gameKindIndex(kind)].Add(1)
	if !startedAt.IsZero() {
		service.metrics.durationNanoseconds.Add(uint64(time.Since(startedAt)))
	}
	entries := make([]Score, 0, len(snapshot.Teams))
	for playerID, team := range snapshot.Teams {
		entries = append(entries, Score{RoomID: active.ID(), Kind: kind, StartedAt: startedAt, PlayerID: playerID, Team: team, Score: snapshot.Scores[playerID] - snapshot.WiredScores[playerID], TeamScore: snapshot.TeamScores[team]})
		service.progress(ctx, playerID, "game.played", 1)
		service.progress(ctx, playerID, playedKey(kind), 1)
	}
	service.progressWinners(ctx, snapshot, kind)
	service.progress(ctx, active.Snapshot().OwnerPlayerID, "game.authored", 1)
	if service.scores != nil && len(entries) > 0 {
		return service.scores.Save(ctx, entries)
	}
	return nil
}

// progressWinners publishes winner progress only for a sole highest team.
func (service *Service) progressWinners(ctx context.Context, snapshot wiredgame.State, kind string) {
	winner := winningTeam(snapshot)
	for playerID, team := range snapshot.Teams {
		if team == winner && winner > 0 {
			service.progress(ctx, playerID, winnerKey(kind), 1)
		}
	}
}

// winningTeam returns the sole highest-scoring participating team.
func winningTeam(snapshot wiredgame.State) int32 {
	bestTeam, bestScore, tied := int32(0), int64(0), false
	seen := [5]bool{}
	for _, team := range snapshot.Teams {
		if team < 1 || team > 4 || seen[team] {
			continue
		}
		seen[team] = true
		score := snapshot.TeamScores[team]
		if bestTeam == 0 || score > bestScore {
			bestTeam, bestScore, tied = team, score, false
		} else if score == bestScore {
			tied = true
		}
	}
	if tied {
		return 0
	}
	return bestTeam
}

// freezeMatchOver reports last-team-standing only after multiple teams participated.
func freezeMatchOver(players map[int64]*freeze.Player) bool {
	teams, alive := [5]bool{}, [5]bool{}
	teamCount, aliveCount := 0, 0
	for _, player := range players {
		team := player.Team
		if team < 1 || team > 4 {
			continue
		}
		if !teams[team] {
			teams[team], teamCount = true, teamCount+1
		}
		if player.Alive() && !alive[team] {
			alive[team], aliveCount = true, aliveCount+1
		}
	}
	return teamCount > 1 && aliveCount <= 1
}

// winnerKey maps game kinds to winner achievement triggers.
func winnerKey(kind string) string {
	if kind == "banzai" {
		return "game.banzai.won"
	}
	if kind == "freeze" {
		return "game.freeze.won"
	}
	return ""
}

// playedKey maps game kinds to existing achievement triggers.
func playedKey(kind string) string {
	switch kind {
	case "banzai":
		return "game.battleball.played"
	case "freeze":
		return "game.freeze.played"
	case "football":
		return "game.football.played"
	case "tag":
		return "game.tag.played"
	default:
		return ""
	}
}

// increaseTimer selects the next definition-backed duration.
func (service *Service) increaseTimer(ctx context.Context, request UseRequest) error {
	service.mutex.Lock()
	state := service.stateLocked(request.Room)
	state.timerItemID = request.Item.ID
	remaining := int(state.timer.Increase() / time.Second)
	service.mutex.Unlock()
	return service.projectState(ctx, request.Room, request.Item.ID, remaining)
}

// timerSteps parses the first placed timer's positive comma-separated durations.
func timerSteps(active *roomlive.Room) []time.Duration {
	items := active.FurnitureByInteraction("game_timer")
	if len(items) == 0 {
		return nil
	}
	parts := strings.Split(items[0].Definition.CustomParams, ",")
	steps := make([]time.Duration, 0, len(parts))
	for _, part := range parts {
		seconds, err := strconv.Atoi(strings.TrimSpace(part))
		if err == nil && seconds > 0 {
			steps = append(steps, time.Duration(seconds)*time.Second)
		}
	}
	return steps
}

// kickFootball queues one furniture-click kick from the actor toward the ball.
func (service *Service) kickFootball(request UseRequest) error {
	snapshot, found := service.wired.Snapshot(request.Room.ID())
	unit, present := request.Room.Unit(request.PlayerID)
	if !found || !snapshot.Running || !present || !pointsAdjacent(unit.Position.Point, request.Item.Point) {
		return nil
	}
	service.queueFootball(request.Room, request.Item.ID, rotationBetween(unit.Position.Point, request.Item.Point), request.PlayerID)
	return nil
}
