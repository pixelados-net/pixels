package cleanup

import (
	"context"
	"testing"
	"time"

	cameraconfig "github.com/niflaot/pixels/internal/realm/camera/config"
	camerametrics "github.com/niflaot/pixels/internal/realm/camera/observability"
	camerarecord "github.com/niflaot/pixels/internal/realm/camera/record"
	"go.uber.org/zap"
)

// TestSchedulerStartsSweepsAndStops verifies one aggregate loop lifecycle.
func TestSchedulerStartsSweepsAndStops(t *testing.T) {
	store := &schedulerStore{called: make(chan struct{}, 1)}
	config := cameraconfig.Config{PendingRetention: time.Hour, SupersededRetention: time.Hour, CleanupInterval: time.Millisecond, CleanupRetry: time.Hour, CleanupBatchSize: 1}
	service := New(config, store, &cleanupStorage{}, camerametrics.New(), zap.NewNop())
	scheduler := NewScheduler(config, service, zap.NewNop())
	if err := scheduler.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	select {
	case <-store.called:
	case <-time.After(time.Second):
		t.Fatal("scheduler did not execute a sweep")
	}
	stopCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := scheduler.Stop(stopCtx); err != nil {
		t.Fatal(err)
	}
}

// schedulerStore observes one scheduler cleanup claim.
type schedulerStore struct {
	// called receives one signal for each claim.
	called chan struct{}
}

// ClaimCleanup records a scheduler sweep.
func (store *schedulerStore) ClaimCleanup(context.Context, time.Time, time.Time, time.Time, int) ([]camerarecord.CleanupCandidate, error) {
	select {
	case store.called <- struct{}{}:
	default:
	}
	return nil, nil
}

// MarkDeleted reports no finalized candidate.
func (*schedulerStore) MarkDeleted(context.Context, int64, time.Time) (bool, error) {
	return false, nil
}
