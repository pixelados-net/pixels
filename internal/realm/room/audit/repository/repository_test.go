package repository

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	auditmodel "github.com/niflaot/pixels/internal/realm/room/audit/model"
	moderationmodel "github.com/niflaot/pixels/internal/realm/room/moderation/model"
)

// executorForTest stores deterministic audit query results.
type executorForTest struct {
	// rows stores the next query rows.
	rows pgx.Rows
	// execs counts append-only writes.
	execs int
}

// Exec records one append-only write.
func (executor *executorForTest) Exec(context.Context, string, ...any) (pgconn.CommandTag, error) {
	executor.execs++
	return pgconn.NewCommandTag("INSERT 0 1"), nil
}

// Query returns configured audit rows.
func (executor *executorForTest) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return executor.rows, nil
}

// QueryRow returns no row.
func (*executorForTest) QueryRow(context.Context, string, ...any) pgx.Row { return rowForTest{} }

// rowForTest is an unused query row.
type rowForTest struct{}

// Scan returns no error.
func (rowForTest) Scan(...any) error { return nil }

// auditRowsForTest scans one rights or moderation row.
type auditRowsForTest struct {
	// moderation selects moderation shape.
	moderation bool
	// remaining stores unread rows.
	remaining int
}

// Close closes rows.
func (*auditRowsForTest) Close() {}

// Err returns no row error.
func (*auditRowsForTest) Err() error { return nil }

// CommandTag returns an empty tag.
func (*auditRowsForTest) CommandTag() pgconn.CommandTag { return pgconn.CommandTag{} }

// FieldDescriptions returns no descriptions.
func (*auditRowsForTest) FieldDescriptions() []pgconn.FieldDescription { return nil }

// Next advances one row.
func (rows *auditRowsForTest) Next() bool {
	if rows.remaining == 0 {
		return false
	}
	rows.remaining--
	return true
}

// Scan writes one audit row.
func (rows *auditRowsForTest) Scan(destinations ...any) error {
	now := time.Unix(1000, 0)
	actorID := int64(1)
	*destinations[0].(*int64) = 5
	*destinations[1].(*int64) = 9
	if rows.moderation {
		*destinations[2].(*int64) = 2
		*destinations[3].(*string) = "player"
		*destinations[4].(**int64) = &actorID
		*destinations[5].(*moderationmodel.Action) = moderationmodel.ActionKick
		*destinations[6].(**int64) = nil
		*destinations[7].(**time.Time) = nil
		*destinations[8].(*time.Time) = now
		return nil
	}
	*destinations[2].(*int64) = 2
	*destinations[3].(*string) = "player"
	*destinations[4].(**int64) = &actorID
	*destinations[5].(*auditmodel.RightsAction) = auditmodel.RightsGranted
	*destinations[6].(*time.Time) = now
	return nil
}

// Values returns no raw values.
func (*auditRowsForTest) Values() ([]any, error) { return nil, nil }

// RawValues returns no raw values.
func (*auditRowsForTest) RawValues() [][]byte { return nil }

// Conn returns no physical connection.
func (*auditRowsForTest) Conn() *pgx.Conn { return nil }

// TestRepositoryAppendsAuditRows verifies append-only writes.
func TestRepositoryAppendsAuditRows(t *testing.T) {
	executor := &executorForTest{}
	repository := &Repository{executor: executor}
	ctx := context.Background()
	if err := repository.InsertRights(ctx, auditmodel.RightsAudit{RoomID: 9, PlayerID: 2}); err != nil {
		t.Fatalf("insert rights: %v", err)
	}
	if err := repository.InsertModeration(ctx, auditmodel.ModerationAction{RoomID: 9, TargetPlayerID: 2}); err != nil {
		t.Fatalf("insert moderation: %v", err)
	}
	if executor.execs != 2 {
		t.Fatalf("expected two inserts, got %d", executor.execs)
	}
}

// TestRepositoryReadsAuditRows verifies rights and moderation query scanning.
func TestRepositoryReadsAuditRows(t *testing.T) {
	executor := &executorForTest{rows: &auditRowsForTest{remaining: 1}}
	repository := &Repository{executor: executor}
	rights, err := repository.RightsHistory(context.Background(), Query{Limit: 50})
	if err != nil || len(rights) != 1 || rights[0].Action != auditmodel.RightsGranted {
		t.Fatalf("rights=%#v err=%v", rights, err)
	}
	executor.rows = &auditRowsForTest{moderation: true, remaining: 1}
	actions, err := repository.ModerationHistory(context.Background(), Query{Limit: 50, ActionTypes: []moderationmodel.Action{moderationmodel.ActionKick}})
	if err != nil || len(actions) != 1 || actions[0].Action != moderationmodel.ActionKick {
		t.Fatalf("actions=%#v err=%v", actions, err)
	}
}
