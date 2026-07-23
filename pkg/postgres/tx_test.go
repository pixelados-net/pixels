package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TestWithinTxReturnsBeginError verifies transaction begin failures are wrapped.
func TestWithinTxReturnsBeginError(t *testing.T) {
	config := testConfig()
	config.Port = 1
	config.ConnectTimeout = time.Millisecond

	poolConfig, err := pgxpool.ParseConfig(config.DSN())
	if err != nil {
		t.Fatalf("parse pool config: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		t.Fatalf("new pool: %v", err)
	}
	t.Cleanup(pool.Close)

	err = WithinTx(context.Background(), pool, func(context.Context, pgx.Tx) error {
		t.Fatal("unexpected transaction callback")

		return nil
	})
	if err == nil {
		t.Fatal("expected begin error")
	}

	if !strings.Contains(err.Error(), "begin postgres transaction") {
		t.Fatalf("expected wrapped begin error, got %v", err)
	}
}

// TestRollbackReturnsCause verifies rollback preserves the original error.
func TestRollbackReturnsCause(t *testing.T) {
	expected := errors.New("write failed")

	err := rollback(context.Background(), &fakeRollbacker{}, expected)
	if !errors.Is(err, expected) {
		t.Fatalf("expected cause, got %v", err)
	}
}

// TestFinishTxCommitsWithoutCause verifies successful work commits.
func TestFinishTxCommitsWithoutCause(t *testing.T) {
	tx := &fakeTransaction{}

	if err := finishTx(context.Background(), tx, nil); err != nil {
		t.Fatalf("finish transaction: %v", err)
	}

	if !tx.committed {
		t.Fatal("expected commit")
	}
}

// TestFinishTxWrapsCommitError verifies commit errors are wrapped.
func TestFinishTxWrapsCommitError(t *testing.T) {
	expected := errors.New("commit failed")

	err := finishTx(context.Background(), &fakeTransaction{commitErr: expected}, nil)
	if !errors.Is(err, expected) {
		t.Fatalf("expected commit error, got %v", err)
	}
}

// TestFinishTxRollsBackCause verifies failed work rolls back.
func TestFinishTxRollsBackCause(t *testing.T) {
	expected := errors.New("work failed")
	tx := &fakeTransaction{}

	err := finishTx(context.Background(), tx, expected)
	if !errors.Is(err, expected) {
		t.Fatalf("expected work error, got %v", err)
	}

	if !tx.rolledBack {
		t.Fatal("expected rollback")
	}
}

// TestRollbackWrapsRollbackError verifies rollback errors include the cause.
func TestRollbackWrapsRollbackError(t *testing.T) {
	cause := errors.New("write failed")
	rollbackError := errors.New("network failed")

	err := rollback(context.Background(), &fakeRollbacker{err: rollbackError}, cause)
	if !errors.Is(err, cause) {
		t.Fatalf("expected cause wrapping, got %v", err)
	}

	if !errors.Is(err, rollbackError) {
		t.Fatalf("expected rollback wrapping, got %v", err)
	}
}

// fakeRollbacker records transaction rollback for tests.
type fakeRollbacker struct {
	// err is the error returned by Rollback.
	err error
}

// Rollback rolls back a transaction for tests.
func (rollbacker *fakeRollbacker) Rollback(context.Context) error {
	return rollbacker.err
}

// fakeTransaction records transaction completion for tests.
type fakeTransaction struct {
	// commitErr is the error returned by Commit.
	commitErr error

	// rollbackErr is the error returned by Rollback.
	rollbackErr error

	// committed reports whether Commit was called.
	committed bool

	// rolledBack reports whether Rollback was called.
	rolledBack bool
}

// Commit commits a transaction for tests.
func (tx *fakeTransaction) Commit(context.Context) error {
	tx.committed = true

	return tx.commitErr
}

// Rollback rolls back a transaction for tests.
func (tx *fakeTransaction) Rollback(context.Context) error {
	tx.rolledBack = true

	return tx.rollbackErr
}
