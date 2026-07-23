package observability

import "testing"

// TestMetricsSnapshot verifies bounded counter projection.
func TestMetricsSnapshot(t *testing.T) {
	metrics := New()
	metrics.Capture(false, 10)
	metrics.Capture(true, 5)
	metrics.UploadFailed(UploadFailureStorage)
	metrics.UploadFailed(UploadFailureTimeout)
	metrics.UploadFailed(UploadFailureTooLarge)
	metrics.UploadFailed(UploadFailureReceipt)
	metrics.Purchased()
	metrics.Published()
	metrics.Reported()
	metrics.Cleanup(true)
	metrics.Cleanup(false)
	snapshot := metrics.Snapshot()
	if snapshot.Photos != 1 || snapshot.Thumbnails != 1 || snapshot.UploadFailures != 4 || snapshot.UploadStorageFailures != 1 || snapshot.UploadTimeouts != 1 || snapshot.UploadTooLarge != 1 || snapshot.UploadReceiptFailures != 1 || snapshot.Purchases != 1 || snapshot.Publications != 1 || snapshot.Reports != 1 || snapshot.UploadedBytes != 15 || snapshot.CleanupDeleted != 1 || snapshot.CleanupFailures != 1 {
		t.Fatalf("unexpected metrics: %+v", snapshot)
	}
}

// BenchmarkMetricsCapture measures the lock-free camera hot counter.
func BenchmarkMetricsCapture(b *testing.B) {
	metrics := New()
	b.ReportAllocs()
	for range b.N {
		metrics.Capture(false, 512)
	}
}
