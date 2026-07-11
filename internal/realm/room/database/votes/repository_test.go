package votes

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	roomvotes "github.com/niflaot/pixels/internal/realm/room/control/votes"
)

// TestCastIncrementsOnlyNewVotes verifies atomic idempotency branches.
func TestCastIncrementsOnlyNewVotes(t *testing.T) {
	tests := []struct {
		name     string
		tag      pgconn.CommandTag
		score    int
		inserted bool
	}{
		{name: "new vote", tag: pgconn.NewCommandTag("INSERT 0 1"), score: 8, inserted: true},
		{name: "duplicate vote", tag: pgconn.NewCommandTag("INSERT 0 0"), score: 7},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			executor := &fakeExecutor{tag: test.tag, row: fakeRow{values: []any{test.score}}}
			result, err := New(executor).Cast(context.Background(), 4, 2)
			if err != nil || result.Score != test.score || result.Inserted != test.inserted || executor.execs != 1 {
				t.Fatalf("result=%+v execs=%d err=%v", result, executor.execs, err)
			}
		})
	}
}

// TestExistingAvoidsEmptyQueries verifies empty-room query avoidance.
func TestExistingAvoidsEmptyQueries(t *testing.T) {
	executor := &fakeExecutor{}
	voters, err := New(executor).Existing(context.Background(), 4, nil)
	if err != nil || len(voters) != 0 || executor.queries != 0 {
		t.Fatalf("voters=%v queries=%d err=%v", voters, executor.queries, err)
	}
}

// TestExistingMapsActiveVoters verifies batched eligibility reads.
func TestExistingMapsActiveVoters(t *testing.T) {
	executor := &fakeExecutor{rows: &fakeRows{values: [][]any{{int64(2)}, {int64(4)}}}}
	voters, err := New(executor).Existing(context.Background(), 4, []int64{2, 3, 4})
	if err != nil || len(voters) != 2 {
		t.Fatalf("voters=%v err=%v", voters, err)
	}
}

// TestReadOperationsMapVotes verifies state and administrative reads.
func TestReadOperationsMapVotes(t *testing.T) {
	now := time.Now().UTC()
	executor := &fakeExecutor{row: fakeRow{values: []any{true}}, rows: &fakeRows{values: [][]any{{int64(4), int64(2), now}}}}
	repository := New(executor)
	voted, err := repository.HasVote(context.Background(), 4, 2)
	if err != nil || !voted {
		t.Fatalf("has vote=%v err=%v", voted, err)
	}
	items, err := repository.List(context.Background(), roomvotes.Query{RoomID: 4, Limit: 10})
	if err != nil || len(items) != 1 || items[0].PlayerID != 2 || !items[0].CreatedAt.Equal(now) {
		t.Fatalf("items=%+v err=%v", items, err)
	}
}

// TestRepositoryPropagatesDatabaseFailures verifies meaningful persistence errors.
func TestRepositoryPropagatesDatabaseFailures(t *testing.T) {
	cause := errors.New("database unavailable")
	if _, err := New(&fakeExecutor{execErr: cause}).Cast(context.Background(), 4, 2); !errors.Is(err, cause) {
		t.Fatalf("cast error=%v", err)
	}
	if _, err := New(&fakeExecutor{row: fakeRow{err: cause}}).HasVote(context.Background(), 4, 2); !errors.Is(err, cause) {
		t.Fatalf("has vote error=%v", err)
	}
	if _, err := New(&fakeExecutor{queryErr: cause}).Existing(context.Background(), 4, []int64{2}); !errors.Is(err, cause) {
		t.Fatalf("existing error=%v", err)
	}
	if _, err := New(&fakeExecutor{queryErr: cause}).List(context.Background(), roomvotes.Query{RoomID: 4, Limit: 1}); !errors.Is(err, cause) {
		t.Fatalf("list error=%v", err)
	}
	if _, err := New(&fakeExecutor{rows: &fakeRows{err: cause}}).List(context.Background(), roomvotes.Query{RoomID: 4, Limit: 1}); !errors.Is(err, cause) {
		t.Fatalf("list iteration error=%v", err)
	}
}
