package database

import (
	"context"
	"time"

	sanctionrecord "github.com/niflaot/pixels/internal/realm/sanction/record"
)

// QueueAlert persists one offline warning.
func (repository *Repository) QueueAlert(ctx context.Context, alert sanctionrecord.Alert) error {
	_, err := repository.executor(ctx).Exec(ctx, `insert into pending_alerts(player_id,punishment_id,message) values($1,$2,$3)`, alert.PlayerID, alert.PunishmentID, alert.Message)
	return err
}

// PendingAlerts returns warnings that still require successful delivery.
func (repository *Repository) PendingAlerts(ctx context.Context, playerID int64, limit int32) ([]sanctionrecord.Alert, error) {
	rows, err := repository.executor(ctx).Query(ctx, `select id,player_id,punishment_id,message from pending_alerts where player_id=$1 and delivered_at is null order by id limit $2`, playerID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := make([]sanctionrecord.Alert, 0)
	for rows.Next() {
		var value sanctionrecord.Alert
		if err = rows.Scan(&value.ID, &value.PlayerID, &value.PunishmentID, &value.Message); err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

// MarkAlertDelivered acknowledges one warning after its packet was sent.
func (repository *Repository) MarkAlertDelivered(ctx context.Context, id int64, deliveredAt time.Time) error {
	_, err := repository.executor(ctx).Exec(ctx, `update pending_alerts set delivered_at=$2 where id=$1 and delivered_at is null`, id, deliveredAt)
	return err
}

var _ sanctionrecord.Store = (*Repository)(nil)
