package game

import (
	"context"
	"encoding/json"
	"math"
	"time"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
)

// boardProjection stores durable JSON consumed as Nitro highscore stuff-data.
type boardProjection struct {
	// State stores the legacy visual state.
	State string `json:"state"`
	// ScoreType stores Nitro's board grouping code.
	ScoreType int32 `json:"score_type"`
	// ClearType stores Nitro's UTC rollover code.
	ClearType int32 `json:"clear_type"`
	// Entries stores bounded ranked rows.
	Entries []boardEntry `json:"entries"`
}

// boardEntry stores one protocol-safe ranked row.
type boardEntry struct {
	// Score stores the clamped Nitro integer score.
	Score int32 `json:"score"`
	// Users stores current participant usernames.
	Users []string `json:"users"`
}

// projectBoards persists and broadcasts every active room board.
func (coordinator *Coordinator) projectBoards(ctx context.Context, active *roomlive.Room, state State, winners map[int32]struct{}) error {
	for _, board := range active.FurnitureByInteraction("wf_highscore") {
		mode, period, scoreType, clearType, valid := boardKind(board.Definition.SpriteID)
		if !valid {
			continue
		}
		start := periodStart(coordinator.now().UTC(), period)
		entries, err := coordinator.highscores.RecordAndList(ctx, board.ID, mode, period, start, resultsFor(mode, state, winners), coordinator.highscoreTop)
		if err != nil {
			return err
		}
		encoded, err := json.Marshal(projectedBoard(scoreType, clearType, entries))
		if err != nil {
			return err
		}
		if err = coordinator.updateItem(ctx, active, board, string(encoded)); err != nil {
			return err
		}
	}
	return nil
}

// boardKind derives canonical mode and period from audited Arcturus sprites.
func boardKind(spriteID int) (record.HighscoreMode, record.HighscorePeriod, int32, int32, bool) {
	var mode record.HighscoreMode
	var scoreType int32
	var offset int
	switch {
	case spriteID >= 5044 && spriteID <= 5047:
		mode, scoreType, offset = record.HighscoreClassic, 2, spriteID-5044
	case spriteID >= 5051 && spriteID <= 5054:
		mode, scoreType, offset = record.HighscorePerTeam, 0, spriteID-5051
	case spriteID >= 5057 && spriteID <= 5060:
		mode, scoreType, offset = record.HighscoreMostWins, 1, spriteID-5057
	default:
		return "", "", 0, 0, false
	}
	periods := [...]record.HighscorePeriod{record.HighscoreAllTime, record.HighscoreDaily, record.HighscoreWeekly, record.HighscoreMonthly}
	return mode, periods[offset], scoreType, int32(offset), true
}

// periodStart returns the normalized UTC window start or nil for all-time.
func periodStart(now time.Time, period record.HighscorePeriod) *time.Time {
	var start time.Time
	switch period {
	case record.HighscoreDaily:
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	case record.HighscoreWeekly:
		days := (int(now.Weekday()) + 6) % 7
		start = time.Date(now.Year(), now.Month(), now.Day()-days, 0, 0, 0, 0, time.UTC)
	case record.HighscoreMonthly:
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	default:
		return nil
	}
	return &start
}

// projectedBoard maps durable rankings to bounded protocol data.
func projectedBoard(scoreType int32, clearType int32, entries []record.HighscoreEntry) boardProjection {
	result := boardProjection{State: "0", ScoreType: scoreType, ClearType: clearType, Entries: make([]boardEntry, 0, len(entries))}
	for _, entry := range entries {
		score := entry.Score
		if score > math.MaxInt32 {
			score = math.MaxInt32
		}
		if score < math.MinInt32 {
			score = math.MinInt32
		}
		result.Entries = append(result.Entries, boardEntry{Score: int32(score), Users: entry.Usernames})
	}
	return result
}
