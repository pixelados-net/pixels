package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

// TestNewPoolCreatesPool verifies pool construction without opening a connection.
func TestNewPoolCreatesPool(t *testing.T) {
	lifecycle := fxtest.NewLifecycle(t)
	pool, err := NewPool(lifecycle, testConfig(), zap.NewNop())
	if err != nil {
		t.Fatalf("new pool: %v", err)
	}
	t.Cleanup(pool.Close)

	if pool == nil {
		t.Fatal("expected pool")
	}
}

// TestApplyPoolConfig verifies pool limits are copied from configuration.
func TestApplyPoolConfig(t *testing.T) {
	poolConfig, err := pgxpool.ParseConfig(testConfig().DSN())
	if err != nil {
		t.Fatalf("parse pool config: %v", err)
	}

	config := testConfig()
	applyPoolConfig(poolConfig, config)

	if poolConfig.MaxConns != config.MaxConns {
		t.Fatalf("expected max conns %d, got %d", config.MaxConns, poolConfig.MaxConns)
	}

	if poolConfig.AfterConnect == nil {
		t.Fatal("expected after connect hook")
	}
}

// TestStartPoolReturnsPingError verifies startup ping failures are wrapped.
func TestStartPoolReturnsPingError(t *testing.T) {
	expected := errors.New("offline")

	err := startPool(context.Background(), &fakePinger{err: expected}, testConfig(), zap.NewNop())
	if err == nil {
		t.Fatal("expected startup error")
	}

	if !strings.Contains(err.Error(), "ping postgres") {
		t.Fatalf("expected wrapped ping error, got %v", err)
	}
}

// TestStartPoolSucceeds verifies successful startup ping.
func TestStartPoolSucceeds(t *testing.T) {
	err := startPool(context.Background(), &fakePinger{}, testConfig(), zap.NewNop())
	if err != nil {
		t.Fatalf("start pool: %v", err)
	}
}

// testConfig returns PostgreSQL config for tests.
func testConfig() Config {
	return Config{
		Host:             "localhost",
		Port:             5432,
		Database:         "pixels",
		User:             "pixels",
		Password:         "pixels",
		SSLMode:          "disable",
		MaxConns:         10,
		MinConns:         1,
		ConnectTimeout:   5 * time.Second,
		StatementTimeout: 5 * time.Second,
		HealthTimeout:    2 * time.Second,
	}
}
