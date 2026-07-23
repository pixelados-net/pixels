package guardian

import (
	"context"
	"testing"
	"time"

	moderationconfig "github.com/niflaot/pixels/internal/realm/moderation/config"
	moderationrecord "github.com/niflaot/pixels/internal/realm/moderation/record"
)

// poolForTest supplies deterministic reviewers.
type poolForTest struct{ ids []int64 }

// Guardians returns configured ids.
func (pool poolForTest) Guardians(_ int64, limit int) []int64 {
	if limit > len(pool.ids) {
		limit = len(pool.ids)
	}
	return append([]int64(nil), pool.ids[:limit]...)
}

// TestStrictMajorityAndTie verifies verdict aggregation.
func TestStrictMajorityAndTie(t *testing.T) {
	manager := New(moderationconfig.Config{GuardianCount: 3, GuardianVoteWindow: time.Minute}, poolForTest{ids: []int64{2, 3, 4}})
	if _, err := manager.Create(context.Background(), 1, 9, nil); err != nil {
		t.Fatal(err)
	}
	for _, id := range []int64{2, 3, 4} {
		if _, err := manager.Decide(id, true); err != nil {
			t.Fatal(err)
		}
	}
	for _, vote := range []struct {
		id    int64
		value Verdict
	}{{2, VerdictBad}, {3, VerdictBad}, {4, VerdictAcceptable}} {
		ticket, _, err := manager.Vote(vote.id, vote.value)
		if err != nil {
			t.Fatal(err)
		}
		if vote.id == 4 && ticket.Result != VerdictBad {
			t.Fatalf("result=%d", ticket.Result)
		}
	}
}

// TestReviewersCannotOverlapOrReviewTheTarget verifies ticket participant isolation.
func TestReviewersCannotOverlapOrReviewTheTarget(t *testing.T) {
	manager := New(moderationconfig.Config{GuardianCount: 2, GuardianVoteWindow: time.Minute}, poolForTest{ids: []int64{9, 2, 3, 4}})
	first, err := manager.Create(context.Background(), 1, 9, nil)
	if err != nil {
		t.Fatal(err)
	}
	if first.Reviewers[9] != nil {
		t.Fatal("reported player was offered their own review")
	}
	second, err := manager.Create(context.Background(), 8, 7, nil)
	if err != nil {
		t.Fatal(err)
	}
	for id := range first.Reviewers {
		if second.Reviewers[id] != nil {
			t.Fatalf("guardian %d was assigned to overlapping tickets", id)
		}
	}
}

// TestClosedTicketsAreReleased verifies completed reviews do not accumulate in memory.
func TestClosedTicketsAreReleased(t *testing.T) {
	manager := New(moderationconfig.Config{GuardianCount: 1, GuardianVoteWindow: time.Minute}, poolForTest{ids: []int64{2}})
	ticket, err := manager.Create(context.Background(), 1, 9, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = manager.Decide(2, true); err != nil {
		t.Fatal(err)
	}
	if _, _, err = manager.Vote(2, VerdictAcceptable); err != nil {
		t.Fatal(err)
	}
	if manager.tickets[ticket.ID] != nil || manager.byPlayer[1] != 0 || manager.byPlayer[2] != 0 {
		t.Fatal("closed ticket retained participant indexes")
	}
}

// TestGuardianValidation verifies unavailable, duplicate, and unauthorized transitions.
func TestGuardianValidation(t *testing.T) {
	config := moderationconfig.Config{GuardianCount: 2, GuardianVoteWindow: time.Minute}
	manager := New(config, poolForTest{ids: []int64{2}})
	if _, err := manager.Create(context.Background(), 1, 9, nil); err != ErrUnavailable {
		t.Fatalf("unavailable err=%v", err)
	}
	manager = New(config, poolForTest{ids: []int64{2, 3, 4, 5}})
	if _, err := manager.Create(context.Background(), 1, 9, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Create(context.Background(), 1, 8, nil); err != ErrInvalidState {
		t.Fatalf("duplicate err=%v", err)
	}
	if _, err := manager.Decide(4, true); err != ErrInvalidState {
		t.Fatalf("foreign decision err=%v", err)
	}
	if _, _, err := manager.Vote(2, Verdict(99)); err != ErrInvalidState {
		t.Fatalf("invalid vote err=%v", err)
	}
	if _, _, err := manager.Vote(2, VerdictBad); err != ErrInvalidState {
		t.Fatalf("early vote err=%v", err)
	}
}

// TestGuardianExpiryAndDetachReleaseState verifies global ticks and abandonment cleanup.
func TestGuardianExpiryAndDetachReleaseState(t *testing.T) {
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	manager := New(moderationconfig.Config{GuardianCount: 1, GuardianVoteWindow: time.Minute}, poolForTest{ids: []int64{2}})
	manager.now = func() time.Time { return now }
	ticket, err := manager.Create(context.Background(), 1, 9, nil)
	if err != nil {
		t.Fatal(err)
	}
	closed := manager.Tick(context.Background(), now.Add(2*time.Minute))
	if len(closed) != 1 || closed[0].State != StateClosed || manager.tickets[ticket.ID] != nil {
		t.Fatalf("closed=%+v", closed)
	}
	ticket, err = manager.Create(context.Background(), 1, 9, nil)
	if err != nil {
		t.Fatal(err)
	}
	value, found := manager.Detach(2)
	if !found || value.State != StateClosed || manager.tickets[ticket.ID] != nil {
		t.Fatalf("value=%+v found=%v", value, found)
	}
}

// TestAnonymizeAndCloneDetachEvidence verifies aliases and copied mutable state.
func TestAnonymizeAndCloneDetachEvidence(t *testing.T) {
	first, second := int64(7), int64(8)
	entries := []moderationrecord.ChatEntry{{PlayerID: &first, Message: "one"}, {PlayerID: &second, Message: "two"}, {PlayerID: &first, Message: "three"}}
	values := anonymize(entries)
	if values[0].PatternID != "UserA" || values[1].PatternID != "UserB" || values[2].PatternID != "UserA" || values[0].PlayerID != nil {
		t.Fatalf("values=%+v", values)
	}
	ticket := &Ticket{Chatlog: values, Reviewers: map[int64]*Reviewer{2: {PlayerID: 2}}}
	detached := clone(ticket)
	detached.Chatlog[0].Message = "mutated"
	detached.Reviewers[2].Accepted = true
	if ticket.Chatlog[0].Message != "one" || ticket.Reviewers[2].Accepted {
		t.Fatal("clone retained mutable ticket storage")
	}
}

// TestGuardianAccessorsAndHelpers verifies snapshots, completion checks, and stable Redis keys.
func TestGuardianAccessorsAndHelpers(t *testing.T) {
	manager := NewPersistent(moderationconfig.Config{GuardianCount: 1, GuardianVoteWindow: time.Minute}, poolForTest{ids: []int64{2}}, nil, nil)
	if _, err := manager.Create(context.Background(), 1, 9, nil); err != nil {
		t.Fatal(err)
	}
	ticket, found := manager.TicketFor(2)
	if !found || ticket.ReporterPlayerID != 1 {
		t.Fatalf("ticket=%+v found=%v", ticket, found)
	}
	if _, found = manager.TicketFor(8); found {
		t.Fatal("missing participant resolved a ticket")
	}
	if votingComplete(&Ticket{Reviewers: map[int64]*Reviewer{2: {Accepted: true}}}) {
		t.Fatal("missing accepted vote reported complete")
	}
	vote := VerdictBad
	if !votingComplete(&Ticket{Reviewers: map[int64]*Reviewer{2: {Accepted: true, Vote: &vote}}}) {
		t.Fatal("complete accepted vote reported incomplete")
	}
	if guardianIgnoredKey(7) != "moderation:guardian:ignored:7" || guardianExclusionKey(7) != "moderation:guardian:excluded:7" {
		t.Fatal("guardian Redis keys changed")
	}
}
