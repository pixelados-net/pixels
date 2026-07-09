package repository

import (
	"context"
	"errors"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/niflaot/pixels/pkg/postgres"
)

// TestGrantMutatesAndWritesLedger verifies transactional delta and audit behavior.
func TestGrantMutatesAndWritesLedger(t *testing.T) {
	now := time.Now()
	executor := &fakeExecutor{rows: []pgx.Row{
		fakeRow{values: balanceValues(10, 1, now)},
		fakeRow{values: balanceValues(15, 2, now)},
	}}
	repository := testRepository(executor)

	result, err := repository.Grant(context.Background(), Mutation{
		PlayerID: 7, CurrencyType: -1, Amount: 5, Ledger: true, ActorKind: "system",
	})
	if err != nil {
		t.Fatalf("grant: %v", err)
	}
	if result.Balance.Amount != 15 || result.Delta != 5 || len(executor.execs) != 2 {
		t.Fatalf("unexpected result=%#v execs=%d", result, len(executor.execs))
	}
	if delta := executor.execs[1].arguments[2]; delta != int64(5) {
		t.Fatalf("unexpected ledger delta %#v", delta)
	}
}

// TestSetRecordsNegativeDelta verifies absolute corrections audit their real delta.
func TestSetRecordsNegativeDelta(t *testing.T) {
	now := time.Now()
	executor := &fakeExecutor{rows: []pgx.Row{
		fakeRow{values: balanceValues(10, 1, now)},
		fakeRow{values: balanceValues(3, 2, now)},
	}}
	repository := testRepository(executor)

	result, err := repository.Set(context.Background(), Mutation{
		PlayerID: 7, CurrencyType: -1, Amount: 3, Ledger: true, ActorKind: "admin",
	})
	if err != nil {
		t.Fatalf("set: %v", err)
	}
	if result.Delta != -7 {
		t.Fatalf("unexpected result delta %d", result.Delta)
	}
	if delta := executor.execs[1].arguments[2]; delta != int64(-7) {
		t.Fatalf("unexpected ledger delta %#v", delta)
	}
}

// TestGrantRejectsNegativeResult verifies insufficient balances skip persistence.
func TestGrantRejectsNegativeResult(t *testing.T) {
	executor := &fakeExecutor{rows: []pgx.Row{fakeRow{err: pgx.ErrNoRows}}}
	repository := testRepository(executor)

	_, err := repository.Grant(context.Background(), Mutation{
		PlayerID: 7, CurrencyType: 0, Amount: -1, ActorKind: "player",
	})
	if !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("expected insufficient balance, got %v", err)
	}
	if executor.rowIndex != 1 {
		t.Fatalf("expected no upsert query, got %d rows", executor.rowIndex)
	}
}

// TestGrantRejectsBalanceOverflow verifies signed storage cannot wrap.
func TestGrantRejectsBalanceOverflow(t *testing.T) {
	executor := &fakeExecutor{rows: []pgx.Row{
		fakeRow{values: balanceValues(math.MaxInt64, 1, time.Now())},
	}}
	repository := testRepository(executor)

	_, err := repository.Grant(context.Background(), Mutation{
		PlayerID: 7, CurrencyType: -1, Amount: 1, ActorKind: "system",
	})
	if !errors.Is(err, ErrBalanceOverflow) {
		t.Fatalf("expected balance overflow, got %v", err)
	}
}

// TestListBalancesScansRows verifies wallet persistence reads.
func TestListBalancesScansRows(t *testing.T) {
	now := time.Now()
	executor := &fakeExecutor{queryRows: &fakeRows{values: [][]any{balanceValues(12, 1, now)}}}
	repository := &Repository{executor: executor}

	balances, err := repository.ListBalances(context.Background(), 7)
	if err != nil {
		t.Fatalf("list balances: %v", err)
	}
	if len(balances) != 1 || balances[0].Amount != 12 || !strings.Contains(executor.query, "player_currencies") {
		t.Fatalf("unexpected balances=%#v query=%q", balances, executor.query)
	}
}

// TestFindBalanceReturnsFoundAndMissingRecords verifies repository lookup behavior.
func TestFindBalanceReturnsFoundAndMissingRecords(t *testing.T) {
	now := time.Now()
	foundRepository := &Repository{executor: &fakeExecutor{rows: []pgx.Row{
		fakeRow{values: balanceValues(12, 1, now)},
	}}}
	balance, found, err := foundRepository.FindBalance(context.Background(), 7, -1)
	if err != nil || !found || balance.Amount != 12 {
		t.Fatalf("unexpected found balance=%#v found=%v err=%v", balance, found, err)
	}

	missingRepository := &Repository{executor: &fakeExecutor{rows: []pgx.Row{
		fakeRow{err: pgx.ErrNoRows},
	}}}
	_, found, err = missingRepository.FindBalance(context.Background(), 7, -1)
	if err != nil || found {
		t.Fatalf("unexpected missing result found=%v err=%v", found, err)
	}
}

// TestRepositoryWrapsDatabaseErrors verifies infrastructure failures remain explicit.
func TestRepositoryWrapsDatabaseErrors(t *testing.T) {
	if New(nil) == nil {
		t.Fatal("expected repository")
	}
	expected := errors.New("database failed")
	if _, err := (&Repository{executor: &fakeExecutor{queryErr: expected}}).ListBalances(context.Background(), 7); err == nil {
		t.Fatal("expected list error")
	}
	if err := lockBalance(context.Background(), &fakeExecutor{execErr: expected}, 7, -1); err == nil {
		t.Fatal("expected lock error")
	}
	if _, err := upsertBalance(context.Background(), &fakeExecutor{rows: []pgx.Row{fakeRow{err: expected}}}, 7, -1, 10); err == nil {
		t.Fatal("expected upsert error")
	}
	if err := insertLedger(context.Background(), &fakeExecutor{execErr: expected}, ledgerEntry(Mutation{PlayerID: 7, CurrencyType: -1}, 1, 1)); err == nil {
		t.Fatal("expected ledger error")
	}
}

// testRepository creates a repository with a synchronous fake transaction.
func testRepository(executor *fakeExecutor) *Repository {
	return &Repository{
		executor: executor,
		withinTx: func(ctx context.Context, work func(context.Context, postgres.Executor) error) error {
			return work(ctx, executor)
		},
	}
}

// balanceValues returns one scannable balance record.
func balanceValues(amount int64, version int64, updatedAt time.Time) []any {
	return []any{int64(7), int32(-1), amount, updatedAt, version}
}

// execCall stores one fake execution.
type execCall struct {
	// query stores executed SQL.
	query string
	// arguments stores SQL arguments.
	arguments []any
}

// fakeExecutor records repository database calls.
type fakeExecutor struct {
	// rows stores QueryRow responses.
	rows []pgx.Row
	// rowIndex stores the next row response.
	rowIndex int
	// queryRows stores Query results.
	queryRows pgx.Rows
	// query stores the last query.
	query string
	// execs stores Exec calls.
	execs []execCall
	// execErr stores an Exec failure.
	execErr error
	// queryErr stores a Query failure.
	queryErr error
}

// Exec records one SQL execution.
func (executor *fakeExecutor) Exec(_ context.Context, query string, arguments ...any) (pgconn.CommandTag, error) {
	executor.execs = append(executor.execs, execCall{query: query, arguments: arguments})
	return pgconn.CommandTag{}, executor.execErr
}

// Query returns configured rows.
func (executor *fakeExecutor) Query(_ context.Context, query string, _ ...any) (pgx.Rows, error) {
	executor.query = query
	return executor.queryRows, executor.queryErr
}

// QueryRow returns the next configured row.
func (executor *fakeExecutor) QueryRow(context.Context, string, ...any) pgx.Row {
	row := executor.rows[executor.rowIndex]
	executor.rowIndex++
	return row
}
