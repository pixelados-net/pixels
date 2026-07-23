package record

import (
	"context"
	"time"
)

// HighscoreMode identifies how one board groups game results.
type HighscoreMode string

const (
	// HighscoreClassic ranks individual player scores.
	HighscoreClassic HighscoreMode = "classic"
	// HighscorePerTeam ranks one normalized participant composition.
	HighscorePerTeam HighscoreMode = "perteam"
	// HighscoreMostWins ranks accumulated victories.
	HighscoreMostWins HighscoreMode = "mostwin"
)

// HighscorePeriod identifies one UTC board window.
type HighscorePeriod string

const (
	// HighscoreAllTime never rolls over.
	HighscoreAllTime HighscorePeriod = "alltime"
	// HighscoreDaily rolls over at UTC midnight.
	HighscoreDaily HighscorePeriod = "daily"
	// HighscoreWeekly rolls over on UTC Monday.
	HighscoreWeekly HighscorePeriod = "weekly"
	// HighscoreMonthly rolls over on the first UTC day of a month.
	HighscoreMonthly HighscorePeriod = "monthly"
)

// HighscoreResult stores one normalized result submitted at game end.
type HighscoreResult struct {
	// PlayerIDs stores stable sorted participant identifiers.
	PlayerIDs []int64
	// Score stores the game score for classic and team boards.
	Score int64
	// Won reports whether most-wins boards increment this composition.
	Won bool
}

// HighscoreEntry stores one ranked board row.
type HighscoreEntry struct {
	// Score stores the ranked score or victory count.
	Score int64
	// PlayerIDs stores stable participant identifiers.
	PlayerIDs []int64
	// Usernames stores current display names in participant order.
	Usernames []string
}

// HighscoreStore persists game results and returns a bounded ranking.
type HighscoreStore interface {
	// RecordAndList atomically records results and returns the current top rows.
	RecordAndList(context.Context, int64, HighscoreMode, HighscorePeriod, *time.Time, []HighscoreResult, int) ([]HighscoreEntry, error)
}

// HighscoreResetter clears durable board entries.
type HighscoreResetter interface {
	// Reset deletes every period entry for selected boards in one room.
	Reset(context.Context, int64, []int64) (int64, error)
}
