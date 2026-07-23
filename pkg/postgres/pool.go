package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// Pool is the reusable PostgreSQL connection pool.
type Pool = pgxpool.Pool

// NewPool creates a PostgreSQL pool and registers lifecycle hooks.
func NewPool(lifecycle fx.Lifecycle, config Config, logger *zap.Logger) (*Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(config.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	applyPoolConfig(poolConfig, config)

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return startPool(ctx, pool, config, logger)
		},
		OnStop: func(context.Context) error {
			pool.Close()

			return nil
		},
	})

	return pool, nil
}

// applyPoolConfig applies pool limits and session settings.
func applyPoolConfig(poolConfig *pgxpool.Config, config Config) {
	poolConfig.MaxConns = config.MaxConns
	poolConfig.MinConns = config.MinConns
	poolConfig.ConnConfig.ConnectTimeout = config.ConnectTimeout
	poolConfig.AfterConnect = func(ctx context.Context, connection *pgx.Conn) error {
		return setStatementTimeout(ctx, connection, config)
	}
}

// setStatementTimeout applies the configured statement timeout.
func setStatementTimeout(ctx context.Context, connection *pgx.Conn, config Config) error {
	_, err := connection.Exec(ctx, "select set_config('statement_timeout', $1, false)", config.StatementTimeout.String())

	return err
}

// startPool verifies the database connection and logs pool details.
func startPool(ctx context.Context, pinger Pinger, config Config, logger *zap.Logger) error {
	health := Health{Pinger: pinger, Timeout: config.HealthTimeout}
	if err := health.Ping(ctx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}

	logger.Info(
		"postgres connected",
		zap.String("host", config.Host),
		zap.Int("port", config.Port),
		zap.String("database", config.Database),
		zap.String("user", config.User),
		zap.String("ssl_mode", config.SSLMode),
		zap.Int32("min_conns", config.MinConns),
		zap.Int32("max_conns", config.MaxConns),
	)

	return nil
}
