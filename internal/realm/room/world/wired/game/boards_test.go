package game

import (
	"reflect"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
)

// TestBoardKindsCoversTwelveVariants verifies every audited sprite and period mapping.
func TestBoardKindsCoversTwelveVariants(t *testing.T) {
	for _, base := range []int{5044, 5051, 5057} {
		for offset := 0; offset < 4; offset++ {
			_, period, _, clearType, found := boardKind(base + offset)
			if !found || clearType != int32(offset) {
				t.Fatalf("sprite %d missing or clear type=%d", base+offset, clearType)
			}
			want := []record.HighscorePeriod{record.HighscoreAllTime, record.HighscoreDaily, record.HighscoreWeekly, record.HighscoreMonthly}[offset]
			if period != want {
				t.Fatalf("sprite %d period=%s, want %s", base+offset, period, want)
			}
		}
	}
	if _, _, _, _, found := boardKind(9999); found {
		t.Fatal("unknown sprite was accepted")
	}
}

// TestPeriodStartAndResultsAreDeterministic verifies UTC rollover and stable participants.
func TestPeriodStartAndResultsAreDeterministic(t *testing.T) {
	now := time.Date(2026, time.July, 15, 13, 45, 0, 0, time.FixedZone("local", -5*3600))
	weekly := periodStart(now.UTC(), record.HighscoreWeekly)
	if weekly == nil || weekly.Weekday() != time.Monday || weekly.Hour() != 0 {
		t.Fatalf("weekly start=%v", weekly)
	}
	if periodStart(now, record.HighscoreAllTime) != nil {
		t.Fatal("all-time unexpectedly has a period start")
	}
	state := State{Teams: map[int64]int32{9: 2, 3: 1, 5: 2}, Scores: map[int64]int64{9: 4, 3: 8, 5: 7}}
	state.TeamScores[1], state.TeamScores[2] = 8, 11
	winners := winningTeams(state)
	if !reflect.DeepEqual(winners, map[int32]struct{}{2: {}}) {
		t.Fatalf("winners=%v", winners)
	}
	classic := resultsFor(record.HighscoreClassic, state, winners)
	if got := []int64{classic[0].PlayerIDs[0], classic[1].PlayerIDs[0], classic[2].PlayerIDs[0]}; !reflect.DeepEqual(got, []int64{3, 5, 9}) {
		t.Fatalf("classic order=%v", got)
	}
	teams := resultsFor(record.HighscorePerTeam, state, winners)
	if !reflect.DeepEqual(teams[1].PlayerIDs, []int64{5, 9}) || !teams[1].Won {
		t.Fatalf("team result=%+v", teams[1])
	}
}
