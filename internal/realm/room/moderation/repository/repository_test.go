package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// executorForTest stores deterministic moderation results.
type executorForTest struct {
	// tag stores mutation results.
	tag pgconn.CommandTag
	// exists stores active sanction state.
	exists bool
	// rows stores sanction query rows.
	rows *rowsForTest
	// err optionally fails executor operations.
	err error
}

// Exec returns the configured command tag.
func (executor *executorForTest) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	return executor.tag, executor.err
}

// Query returns configured sanction rows.
func (executor *executorForTest) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return executor.rows, executor.err
}

// QueryRow returns one exists row.
func (executor *executorForTest) QueryRow(context.Context, string, ...any) pgx.Row {
	return rowForTest{exists: executor.exists, err: executor.err}
}

// rowForTest scans an exists result.
type rowForTest struct {
	// exists stores the result.
	exists bool
	// err optionally fails scanning.
	err error
}

// Scan copies the exists result.
func (row rowForTest) Scan(destinations ...any) error {
	if row.err != nil {
		return row.err
	}
	*destinations[0].(*bool) = row.exists
	return nil
}

// TestRepositoryWrapsExecutorErrors verifies contextual sanction errors.
func TestRepositoryWrapsExecutorErrors(t *testing.T) {
	expected := errors.New("database unavailable")
	repository := &Repository{executor: &executorForTest{err: expected}}
	ctx := context.Background()
	if err := repository.Mute(ctx, 9, 2, time.Now()); !errors.Is(err, expected) {
		t.Fatalf("expected mute error, got %v", err)
	}
	if _, err := repository.IsBanned(ctx, 9, 2, time.Now()); !errors.Is(err, expected) {
		t.Fatalf("expected ban lookup error, got %v", err)
	}
	if _, err := repository.ListMutes(ctx, 9, time.Now()); !errors.Is(err, expected) {
		t.Fatalf("expected list error, got %v", err)
	}
}

// rowsForTest scans one sanction row.
type rowsForTest struct {
	// remaining stores unread rows.
	remaining int
}

// Close closes rows.
func (*rowsForTest) Close() {}

// Err returns no row error.
func (*rowsForTest) Err() error { return nil }

// CommandTag returns an empty tag.
func (*rowsForTest) CommandTag() pgconn.CommandTag { return pgconn.CommandTag{} }

// FieldDescriptions returns no descriptions.
func (*rowsForTest) FieldDescriptions() []pgconn.FieldDescription { return nil }

// Next advances one row.
func (rows *rowsForTest) Next() bool {
	if rows.remaining == 0 {
		return false
	}
	rows.remaining--
	return true
}

// Scan writes one sanction row.
func (*rowsForTest) Scan(destinations ...any) error {
	now := time.Unix(1000, 0)
	*destinations[0].(*int64) = 9
	*destinations[1].(*int64) = 2
	*destinations[2].(*string) = "Alice"
	*destinations[3].(*time.Time) = now.Add(time.Hour)
	*destinations[4].(*time.Time) = now
	*destinations[5].(*time.Time) = now
	return nil
}

// Values returns no raw values.
func (*rowsForTest) Values() ([]any, error) { return nil, nil }

// RawValues returns no raw values.
func (*rowsForTest) RawValues() [][]byte { return nil }

// Conn returns no physical connection.
func (*rowsForTest) Conn() *pgx.Conn { return nil }

// TestRepositoryMutatesSanctions verifies mute and ban state mutations.
func TestRepositoryMutatesSanctions(t *testing.T) {
	executor := &executorForTest{tag: pgconn.NewCommandTag("UPDATE 1")}
	repository := &Repository{executor: executor, withinTx: func(ctx context.Context, work func(context.Context) error) error { return work(ctx) }}
	ctx := context.Background()
	now := time.Unix(1000, 0)
	if err := repository.Mute(ctx, 9, 2, now.Add(time.Minute)); err != nil {
		t.Fatalf("mute: %v", err)
	}
	if removed, err := repository.Unmute(ctx, 9, 2, now); err != nil || !removed {
		t.Fatalf("unmute removed=%v err=%v", removed, err)
	}
	if err := repository.Ban(ctx, 9, 2, now.Add(time.Hour)); err != nil {
		t.Fatalf("ban: %v", err)
	}
	if removed, err := repository.Unban(ctx, 9, 2, now); err != nil || !removed {
		t.Fatalf("unban removed=%v err=%v", removed, err)
	}
}

// TestRepositoryReadsActiveSanctions verifies indexed state and list reads.
func TestRepositoryReadsActiveSanctions(t *testing.T) {
	executor := &executorForTest{exists: true, rows: &rowsForTest{remaining: 1}}
	repository := &Repository{executor: executor}
	ctx := context.Background()
	if active, err := repository.IsMuted(ctx, 9, 2, time.Now()); err != nil || !active {
		t.Fatalf("muted=%v err=%v", active, err)
	}
	if active, err := repository.IsBanned(ctx, 9, 2, time.Now()); err != nil || !active {
		t.Fatalf("banned=%v err=%v", active, err)
	}
	mutes, err := repository.ListMutes(ctx, 9, time.Now())
	if err != nil || len(mutes) != 1 || mutes[0].PlayerID != 2 {
		t.Fatalf("mutes=%#v err=%v", mutes, err)
	}
	executor.rows = &rowsForTest{remaining: 1}
	bans, err := repository.ListBans(ctx, 9, time.Now())
	if err != nil || len(bans) != 1 || bans[0].Username != "Alice" {
		t.Fatalf("bans=%#v err=%v", bans, err)
	}
}
