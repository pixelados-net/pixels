package record

import (
	"context"
	"time"
)

// Store persists global punishments, escalation policy, and pending warnings.
type Store interface {
	// Insert creates one punishment.
	Insert(context.Context, ApplyParams) (Punishment, error)
	// Find returns one punishment by id.
	Find(context.Context, int64) (Punishment, bool, error)
	// List returns a player's punishment history.
	List(context.Context, int64, int32) ([]Punishment, error)
	// Active returns a player's current sanction projection.
	Active(context.Context, int64, time.Time) (ActiveState, error)
	// Revoke marks one punishment revoked atomically.
	Revoke(context.Context, int64, *int64, time.Time) (Punishment, bool, error)
	// LastEscalation returns the newest escalated punishment and level.
	LastEscalation(context.Context, int64) (Punishment, int32, bool, error)
	// Ladder returns escalation policy ordered by level.
	Ladder(context.Context) ([]LadderEntry, error)
	// ReplaceLadder atomically replaces escalation policy.
	ReplaceLadder(context.Context, []LadderEntry) error
	// QueueAlert persists an offline warning.
	QueueAlert(context.Context, Alert) error
	// PendingAlerts returns undelivered warnings without acknowledging them.
	PendingAlerts(context.Context, int64, int32) ([]Alert, error)
	// MarkAlertDelivered acknowledges one warning after successful delivery.
	MarkAlertDelivered(context.Context, int64, time.Time) error
}
