// Package database implements direct-trade audit persistence.
package database

import (
	"context"
	traderecord "github.com/niflaot/pixels/internal/realm/trade/record"
	"github.com/niflaot/pixels/pkg/postgres"
)

// Repository stores trade audit records.
type Repository struct {
	// pool owns database transactions.
	pool *postgres.Pool
}

// New creates a trade repository.
func New(pool *postgres.Pool) *Repository { return &Repository{pool: pool} }

// executor returns the active scoped transaction.
func (repository *Repository) executor(ctx context.Context) postgres.Executor {
	return postgres.ExecutorFor(ctx, repository.pool)
}

// WithinTransaction runs work atomically.
func (repository *Repository) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	if _, ok := postgres.ScopedExecutor(ctx); ok {
		return work(ctx)
	}
	return postgres.WithinScope(ctx, repository.pool, work)
}

// InsertAudit records one completed trade.
func (repository *Repository) InsertAudit(ctx context.Context, audit traderecord.Audit) error {
	_, err := repository.executor(ctx).Exec(ctx, `insert into trade_audit_logs(room_id,first_player_id,second_player_id,first_ip,second_ip,first_item_ids,second_item_ids,first_redeemable_credits,second_redeemable_credits) values($1,$2,$3,nullif($4,'')::inet,nullif($5,'')::inet,$6,$7,$8,$9)`, audit.RoomID, audit.FirstPlayerID, audit.SecondPlayerID, audit.FirstIP, audit.SecondIP, audit.FirstItemIDs, audit.SecondItemIDs, audit.FirstRedeemableCredits, audit.SecondRedeemableCredits)
	return err
}

// ListAudits returns recent trades involving one player.
func (repository *Repository) ListAudits(ctx context.Context, playerID int64, limit int32) ([]traderecord.Audit, error) {
	rows, err := repository.executor(ctx).Query(ctx, `select id,room_id,first_player_id,second_player_id,coalesce(host(first_ip),''),coalesce(host(second_ip),''),first_item_ids,second_item_ids,first_redeemable_credits,second_redeemable_credits,created_at from trade_audit_logs where first_player_id=$1 or second_player_id=$1 order by created_at desc limit $2`, playerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	audits := make([]traderecord.Audit, 0)
	for rows.Next() {
		var audit traderecord.Audit
		if err := rows.Scan(&audit.ID, &audit.RoomID, &audit.FirstPlayerID, &audit.SecondPlayerID, &audit.FirstIP, &audit.SecondIP, &audit.FirstItemIDs, &audit.SecondItemIDs, &audit.FirstRedeemableCredits, &audit.SecondRedeemableCredits, &audit.CreatedAt); err != nil {
			return nil, err
		}
		audits = append(audits, audit)
	}
	return audits, rows.Err()
}

var _ traderecord.Store = (*Repository)(nil)
