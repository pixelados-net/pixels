package record

import (
	"context"
	"time"
)

// Store persists camera captures, publications, policy, and audit state.
type Store interface {
	// WithinTransaction runs work in one shared PostgreSQL transaction.
	WithinTransaction(context.Context, func(context.Context) error) error
	// CreateCapture inserts one uploaded artifact.
	CreateCapture(context.Context, Capture) (Capture, error)
	// ActiveCapture locks and returns the current reusable player photo.
	ActiveCapture(context.Context, int64) (Capture, bool, error)
	// LatestCaptureAt returns the latest capture time for one player and kind.
	LatestCaptureAt(context.Context, int64, Kind) (time.Time, bool, error)
	// AttachPurchase links one furniture copy to its source capture.
	AttachPurchase(context.Context, int64, int64) error
	// Settings returns the singleton operational policy.
	Settings(context.Context) (Settings, error)
	// UpdateSettings applies one optimistic settings replacement.
	UpdateSettings(context.Context, Settings, int64) (Settings, bool, error)
	// PublishCooldown returns one player's last publication time.
	PublishCooldown(context.Context, int64) (time.Time, bool, error)
	// SetPublishCooldown upserts one player's publication time.
	SetPublishCooldown(context.Context, int64, time.Time) error
	// CreatePublication inserts one public gallery entry.
	CreatePublication(context.Context, Capture) (Publication, error)
	// PublicationByCapture returns an existing capture publication.
	PublicationByCapture(context.Context, int64) (Publication, bool, error)
	// Publications lists public gallery entries.
	Publications(context.Context, int, int, bool) ([]Publication, error)
	// RemovePublication soft-removes one gallery entry idempotently.
	RemovePublication(context.Context, int64, string) (bool, error)
	// Captures lists recent captures for one player.
	Captures(context.Context, int64, int) ([]Capture, error)
	// ClaimCleanup atomically claims stale unreferenced photo objects.
	ClaimCleanup(context.Context, time.Time, time.Time, time.Time, int) ([]CleanupCandidate, error)
	// MarkDeleted records successful object storage deletion.
	MarkDeleted(context.Context, int64, time.Time) (bool, error)
	// InsertAudit appends one administrative mutation record.
	InsertAudit(context.Context, Audit) error
}
