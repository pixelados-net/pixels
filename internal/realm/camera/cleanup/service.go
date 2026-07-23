// Package cleanup owns safe reconciliation of abandoned camera objects.
package cleanup

import (
	"context"
	"time"

	cameraconfig "github.com/niflaot/pixels/internal/realm/camera/config"
	camerametrics "github.com/niflaot/pixels/internal/realm/camera/observability"
	camerarecord "github.com/niflaot/pixels/internal/realm/camera/record"
	"go.uber.org/zap"
)

// Storage deletes durable camera objects.
type Storage interface {
	// Delete removes one object by key.
	Delete(context.Context, string) error
}

// Store claims and finalizes cleanup work.
type Store interface {
	// ClaimCleanup atomically claims stale unreferenced photo objects.
	ClaimCleanup(context.Context, time.Time, time.Time, time.Time, int) ([]camerarecord.CleanupCandidate, error)
	// MarkDeleted records successful object storage deletion.
	MarkDeleted(context.Context, int64, time.Time) (bool, error)
}

// Service reconciles stale capture receipts with object storage.
type Service struct {
	// config stores cleanup retention and batch policy.
	config cameraconfig.Config
	// store claims and finalizes durable cleanup work.
	store Store
	// storage removes abandoned objects.
	storage Storage
	// metrics records bounded cleanup outcomes.
	metrics *camerametrics.Metrics
	// log records actionable storage failures.
	log *zap.Logger
	// now supplies deterministic reconciliation time.
	now func() time.Time
}

// New creates an abandoned camera object cleanup service.
func New(config cameraconfig.Config, store Store, storage Storage, metrics *camerametrics.Metrics, log *zap.Logger) *Service {
	return &Service{config: config, store: store, storage: storage, metrics: metrics, log: log, now: time.Now}
}

// Sweep claims and deletes one bounded batch without holding database locks during I/O.
func (service *Service) Sweep(ctx context.Context) (int, error) {
	now := service.now()
	candidates, err := service.store.ClaimCleanup(ctx, now.Add(-service.config.PendingRetention), now.Add(-service.config.SupersededRetention), now.Add(-service.config.CleanupRetry), service.config.CleanupBatchSize)
	if err != nil {
		return 0, err
	}
	deleted := 0
	for _, candidate := range candidates {
		if err = service.delete(ctx, candidate); err != nil {
			service.metrics.Cleanup(false)
			if service.log != nil {
				service.log.Warn("camera object cleanup failed", zap.Int64("capture_id", candidate.CaptureID), zap.String("storage_key", candidate.StorageKey), zap.Error(err))
			}
			continue
		}
		marked, markErr := service.store.MarkDeleted(ctx, candidate.CaptureID, now)
		if markErr != nil {
			return deleted, markErr
		}
		if marked {
			deleted++
			service.metrics.Cleanup(true)
		}
	}
	return deleted, nil
}

// delete removes Nitro's companion wall image before its canonical photo.
func (service *Service) delete(ctx context.Context, candidate camerarecord.CleanupCandidate) error {
	if companionKey, ok := camerarecord.PhotoCompanionKey(candidate.StorageKey); ok {
		if err := service.storage.Delete(ctx, companionKey); err != nil {
			return err
		}
	}
	return service.storage.Delete(ctx, candidate.StorageKey)
}
