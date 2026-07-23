package cleanup

import (
	"context"
	"errors"
	"testing"
	"time"

	cameraconfig "github.com/niflaot/pixels/internal/realm/camera/config"
	camerametrics "github.com/niflaot/pixels/internal/realm/camera/observability"
	camerarecord "github.com/niflaot/pixels/internal/realm/camera/record"
	"go.uber.org/zap"
)

// TestSweepDeletesAndFinalizesBoundedCandidates verifies the complete cleanup workflow.
func TestSweepDeletesAndFinalizesBoundedCandidates(t *testing.T) {
	now := time.Unix(1000, 0)
	store := &cleanupStore{candidates: []camerarecord.CleanupCandidate{{CaptureID: 1, StorageKey: "photos/1/a.png"}, {CaptureID: 2, StorageKey: "photos/1/b.png"}}}
	objects := &cleanupStorage{failKey: "photos/1/b.png"}
	metrics := camerametrics.New()
	config := cameraconfig.Config{PendingRetention: 24 * time.Hour, SupersededRetention: time.Hour, CleanupRetry: 5 * time.Minute, CleanupBatchSize: 10}
	service := New(config, store, objects, metrics, zap.NewNop())
	service.now = func() time.Time { return now }
	deleted, err := service.Sweep(context.Background())
	if err != nil || deleted != 1 {
		t.Fatalf("unexpected cleanup deleted=%d err=%v", deleted, err)
	}
	if !store.pendingBefore.Equal(now.Add(-24*time.Hour)) || !store.supersededBefore.Equal(now.Add(-time.Hour)) || !store.retryBefore.Equal(now.Add(-5*time.Minute)) || store.limit != 10 {
		t.Fatalf("unexpected cleanup claim: %+v", store)
	}
	if len(store.marked) != 1 || store.marked[0] != 1 || len(objects.deleted) != 4 {
		t.Fatalf("unexpected cleanup state store=%+v objects=%+v", store, objects)
	}
	want := []string{"photos/1/a_small.png", "photos/1/a.png", "photos/1/b_small.png", "photos/1/b.png"}
	for index := range want {
		if objects.deleted[index] != want[index] {
			t.Fatalf("unexpected deletion order=%+v", objects.deleted)
		}
	}
	snapshot := metrics.Snapshot()
	if snapshot.CleanupDeleted != 1 || snapshot.CleanupFailures != 1 {
		t.Fatalf("unexpected cleanup metrics: %+v", snapshot)
	}
}

// cleanupStore stores claimed and finalized cleanup observations.
type cleanupStore struct {
	// candidates stores claimed fixtures.
	candidates []camerarecord.CleanupCandidate
	// pendingBefore stores the pending retention cutoff.
	pendingBefore time.Time
	// supersededBefore stores the replacement retention cutoff.
	supersededBefore time.Time
	// retryBefore stores the retry cutoff.
	retryBefore time.Time
	// limit stores the bounded claim size.
	limit int
	// marked stores finalized capture identifiers.
	marked []int64
}

// ClaimCleanup returns configured candidates and records policy cutoffs.
func (store *cleanupStore) ClaimCleanup(_ context.Context, pendingBefore time.Time, supersededBefore time.Time, retryBefore time.Time, limit int) ([]camerarecord.CleanupCandidate, error) {
	store.pendingBefore, store.supersededBefore, store.retryBefore, store.limit = pendingBefore, supersededBefore, retryBefore, limit
	return store.candidates, nil
}

// MarkDeleted records one finalized candidate.
func (store *cleanupStore) MarkDeleted(_ context.Context, captureID int64, _ time.Time) (bool, error) {
	store.marked = append(store.marked, captureID)
	return true, nil
}

// cleanupStorage stores object deletion observations.
type cleanupStorage struct {
	// failKey identifies one injected failure.
	failKey string
	// deleted stores attempted object keys.
	deleted []string
}

// Delete records one object deletion and applies an injected failure.
func (storage *cleanupStorage) Delete(_ context.Context, key string) error {
	storage.deleted = append(storage.deleted, key)
	if key == storage.failKey {
		return errors.New("delete failed")
	}
	return nil
}
