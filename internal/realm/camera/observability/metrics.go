// Package observability provides bounded lock-free camera counters.
package observability

import "sync/atomic"

// UploadFailureReason identifies one bounded upload failure category.
type UploadFailureReason uint8

const (
	// UploadFailureStorage identifies provider or network failures.
	UploadFailureStorage UploadFailureReason = iota
	// UploadFailureTimeout identifies an exhausted upload deadline.
	UploadFailureTimeout
	// UploadFailureTooLarge identifies a policy size rejection.
	UploadFailureTooLarge
	// UploadFailureReceipt identifies persistence after a successful upload.
	UploadFailureReceipt
	// uploadFailureCount bounds the lock-free counter array.
	uploadFailureCount
)

// Snapshot stores one point-in-time camera metric projection.
type Snapshot struct {
	// Photos stores successful photo uploads.
	Photos uint64 `json:"photos"`
	// Thumbnails stores successful thumbnail uploads.
	Thumbnails uint64 `json:"thumbnails"`
	// UploadFailures stores failed storage or receipt writes.
	UploadFailures uint64 `json:"uploadFailures"`
	// UploadStorageFailures stores provider and network failures.
	UploadStorageFailures uint64 `json:"uploadStorageFailures"`
	// UploadTimeouts stores exhausted upload deadlines.
	UploadTimeouts uint64 `json:"uploadTimeouts"`
	// UploadTooLarge stores policy size rejections.
	UploadTooLarge uint64 `json:"uploadTooLarge"`
	// UploadReceiptFailures stores failed durable receipt writes.
	UploadReceiptFailures uint64 `json:"uploadReceiptFailures"`
	// Purchases stores committed photo purchases.
	Purchases uint64 `json:"purchases"`
	// Publications stores committed gallery publications.
	Publications uint64 `json:"publications"`
	// Reports stores accepted photo report attempts.
	Reports uint64 `json:"reports"`
	// UploadedBytes stores successful uploaded bytes.
	UploadedBytes uint64 `json:"uploadedBytes"`
	// CleanupDeleted stores successfully removed abandoned objects.
	CleanupDeleted uint64 `json:"cleanupDeleted"`
	// CleanupFailures stores failed cleanup operations.
	CleanupFailures uint64 `json:"cleanupFailures"`
}

// Metrics stores bounded camera counters without high-cardinality labels.
type Metrics struct {
	// photos counts photo uploads.
	photos atomic.Uint64
	// thumbnails counts thumbnail uploads.
	thumbnails atomic.Uint64
	// uploadFailures counts bounded failure categories.
	uploadFailures [uploadFailureCount]atomic.Uint64
	// purchases counts committed purchases.
	purchases atomic.Uint64
	// publications counts committed publications.
	publications atomic.Uint64
	// reports counts submitted reports.
	reports atomic.Uint64
	// uploadedBytes counts successful image bytes.
	uploadedBytes atomic.Uint64
	// cleanupDeleted counts removed abandoned objects.
	cleanupDeleted atomic.Uint64
	// cleanupFailures counts failed cleanup operations.
	cleanupFailures atomic.Uint64
}

// New creates zeroed camera metrics.
func New() *Metrics { return &Metrics{} }

// Capture records one successful upload.
func (metrics *Metrics) Capture(thumbnail bool, bytes int) {
	if metrics == nil {
		return
	}
	if thumbnail {
		metrics.thumbnails.Add(1)
	} else {
		metrics.photos.Add(1)
	}
	if bytes > 0 {
		metrics.uploadedBytes.Add(uint64(bytes))
	}
}

// UploadFailed records one bounded upload failure category.
func (metrics *Metrics) UploadFailed(reason UploadFailureReason) {
	if metrics != nil && reason < uploadFailureCount {
		metrics.uploadFailures[reason].Add(1)
	}
}

// Purchased records one committed photo purchase.
func (metrics *Metrics) Purchased() {
	if metrics != nil {
		metrics.purchases.Add(1)
	}
}

// Published records one committed gallery publication.
func (metrics *Metrics) Published() {
	if metrics != nil {
		metrics.publications.Add(1)
	}
}

// Reported records one submitted photo report.
func (metrics *Metrics) Reported() {
	if metrics != nil {
		metrics.reports.Add(1)
	}
}

// Cleanup records one abandoned object cleanup outcome.
func (metrics *Metrics) Cleanup(deleted bool) {
	if metrics == nil {
		return
	}
	if deleted {
		metrics.cleanupDeleted.Add(1)
	} else {
		metrics.cleanupFailures.Add(1)
	}
}

// Snapshot returns current bounded counters.
func (metrics *Metrics) Snapshot() Snapshot {
	if metrics == nil {
		return Snapshot{}
	}
	storageFailures := metrics.uploadFailures[UploadFailureStorage].Load()
	timeouts := metrics.uploadFailures[UploadFailureTimeout].Load()
	tooLarge := metrics.uploadFailures[UploadFailureTooLarge].Load()
	receipts := metrics.uploadFailures[UploadFailureReceipt].Load()
	return Snapshot{Photos: metrics.photos.Load(), Thumbnails: metrics.thumbnails.Load(), UploadFailures: storageFailures + timeouts + tooLarge + receipts, UploadStorageFailures: storageFailures, UploadTimeouts: timeouts, UploadTooLarge: tooLarge, UploadReceiptFailures: receipts, Purchases: metrics.purchases.Load(), Publications: metrics.publications.Load(), Reports: metrics.reports.Load(), UploadedBytes: metrics.uploadedBytes.Load(), CleanupDeleted: metrics.cleanupDeleted.Load(), CleanupFailures: metrics.cleanupFailures.Load()}
}
