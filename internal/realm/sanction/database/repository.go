// Package database implements PostgreSQL global sanction persistence.
package database

import (
	"context"
	"fmt"

	sanctionrecord "github.com/niflaot/pixels/internal/realm/sanction/record"
	"github.com/niflaot/pixels/pkg/postgres"
)

// Repository stores punishments and escalation policy.
type Repository struct {
	// pool owns database connections.
	pool *postgres.Pool
}

// New creates a sanction repository.
func New(pool *postgres.Pool) *Repository { return &Repository{pool: pool} }

// executor returns the active transaction or pool.
func (repository *Repository) executor(ctx context.Context) postgres.Executor {
	return postgres.ExecutorFor(ctx, repository.pool)
}

// Insert creates one punishment and returns database timestamps.
func (repository *Repository) Insert(ctx context.Context, params sanctionrecord.ApplyParams) (sanctionrecord.Punishment, error) {
	row := repository.executor(ctx).QueryRow(ctx, `insert into punishments(receiver_player_id,issuer_player_id,issuer_kind,kind,reason,cfh_topic_id,issue_id,source,expires_at) values($1,$2,$3,$4,$5,$6,$7,$8,$9) returning id,receiver_player_id,issuer_player_id,issuer_kind,kind,reason,cfh_topic_id,issue_id,source,issued_at,expires_at,revoked_at,revoked_by_player_id`, params.ReceiverPlayerID, params.IssuerPlayerID, params.IssuerKind, params.Kind, params.Reason, params.CFHTopicID, params.IssueID, params.Source, params.ExpiresAt)
	return scanPunishment(row)
}

// Find returns one punishment by id.
func (repository *Repository) Find(ctx context.Context, id int64) (sanctionrecord.Punishment, bool, error) {
	rows, err := repository.executor(ctx).Query(ctx, punishmentSelect+` where id=$1`, id)
	if err != nil {
		return sanctionrecord.Punishment{}, false, err
	}
	defer rows.Close()
	if !rows.Next() {
		return sanctionrecord.Punishment{}, false, rows.Err()
	}
	value, err := scanPunishment(rows)
	return value, err == nil, err
}

// List returns recent punishment history.
func (repository *Repository) List(ctx context.Context, playerID int64, limit int32) ([]sanctionrecord.Punishment, error) {
	rows, err := repository.executor(ctx).Query(ctx, punishmentSelect+` where receiver_player_id=$1 order by issued_at desc limit $2`, playerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := make([]sanctionrecord.Punishment, 0)
	for rows.Next() {
		value, scanErr := scanPunishment(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

const punishmentSelect = `select id,receiver_player_id,issuer_player_id,issuer_kind,kind,reason,cfh_topic_id,issue_id,source,issued_at,expires_at,revoked_at,revoked_by_player_id from punishments`

// rowScanner scans one database row.
type rowScanner interface{ Scan(...any) error }

// scanPunishment maps one complete punishment row.
func scanPunishment(row rowScanner) (sanctionrecord.Punishment, error) {
	var value sanctionrecord.Punishment
	err := row.Scan(&value.ID, &value.ReceiverPlayerID, &value.IssuerPlayerID, &value.IssuerKind, &value.Kind, &value.Reason, &value.CFHTopicID, &value.IssueID, &value.Source, &value.IssuedAt, &value.ExpiresAt, &value.RevokedAt, &value.RevokedByPlayerID)
	if err != nil {
		return sanctionrecord.Punishment{}, fmt.Errorf("scan punishment: %w", err)
	}
	return value, nil
}
