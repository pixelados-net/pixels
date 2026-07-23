// Package observability owns lock-free low-cardinality social-group telemetry.
package observability

import (
	"sync/atomic"
	"time"
)

// Result classifies one bounded operation result.
type Result uint8

const (
	// Success identifies a completed operation.
	Success Result = iota
	// Rejected identifies an expected domain rejection.
	Rejected
	// Failed identifies an unexpected infrastructure failure.
	Failed
	// Unsupported identifies a safe compatibility-only request.
	Unsupported
	// resultCount stores the bounded result array length.
	resultCount
)

// Family identifies one requested metric family.
type Family uint8

const (
	// Operations stores identity workflow results.
	Operations Family = iota
	// Membership stores membership workflow results.
	Membership
	// SnapshotRefresh stores immutable generation refresh results.
	SnapshotRefresh
	// BadgeCompile stores badge compiler results.
	BadgeCompile
	// HQFurnitureReturn stores headquarters cleanup results.
	HQFurnitureReturn
	// ForumOperations stores forum workflow results.
	ForumOperations
	// ForumUnreadCache stores unread projection results.
	ForumUnreadCache
	// familyCount stores the bounded family array length.
	familyCount
)

// Kind identifies one bounded action within a metric family.
type Kind uint8

const (
	// KindDefault stores single-action family results.
	KindDefault Kind = iota
	// KindCreate stores creation results.
	KindCreate
	// KindUpdate stores identity or settings updates.
	KindUpdate
	// KindBadge stores badge mutations.
	KindBadge
	// KindDeactivate stores group deactivation or invalidation.
	KindDeactivate
	// KindRestore stores retained group restoration.
	KindRestore
	// KindTransfer stores owner or role transfer.
	KindTransfer
	// KindRebind stores home-room replacement.
	KindRebind
	// KindList stores bounded list queries.
	KindList
	// KindJoin stores join or request creation.
	KindJoin
	// KindAccept stores request acceptance.
	KindAccept
	// KindDecline stores request rejection.
	KindDecline
	// KindRemove stores membership removal.
	KindRemove
	// KindFavorite stores favorite changes.
	KindFavorite
	// KindPost stores thread or message creation.
	KindPost
	// KindModerate stores retained forum moderation.
	KindModerate
	// kindCount stores the bounded action array length.
	kindCount
)

// Timing identifies one bounded duration histogram.
type Timing uint8

const (
	// CreateTransaction measures group creation transactions.
	CreateTransaction Timing = iota
	// MemberList measures bounded roster reads.
	MemberList
	// ForumQuery measures bounded forum queries.
	ForumQuery
	// ProjectionFanout measures supported online color projection fan-out.
	ProjectionFanout
	// timingCount stores the bounded timing array length.
	timingCount
)

// ResultCounters stores one counter per Result value.
type ResultCounters struct {
	// Success stores completed calls.
	Success uint64 `json:"success"`
	// Rejected stores expected domain rejections.
	Rejected uint64 `json:"rejected"`
	// Failed stores unexpected failures.
	Failed uint64 `json:"failed"`
	// Unsupported stores compatibility-only calls.
	Unsupported uint64 `json:"unsupported"`
}

// HistogramSnapshot stores fixed cumulative duration observations.
type HistogramSnapshot struct {
	// Buckets stores counts for bounded microsecond and millisecond thresholds.
	Buckets [8]uint64 `json:"buckets"`
	// Count stores all observations.
	Count uint64 `json:"count"`
	// Nanoseconds stores cumulative duration.
	Nanoseconds uint64 `json:"nanoseconds"`
}

// Snapshot contains process-wide group telemetry without unbounded labels.
type Snapshot struct {
	// Families stores counters in Family declaration order.
	Families [familyCount][kindCount]ResultCounters `json:"families"`
	// Timings stores histograms in Timing declaration order.
	Timings [timingCount]HistogramSnapshot `json:"timings"`
}

// Metrics stores lock-free bounded counters and histograms.
type Metrics struct {
	// families stores result counters by metric family.
	families [familyCount][kindCount][resultCount]atomic.Uint64
	// timings stores fixed duration histograms.
	timings [timingCount]histogram
}

// histogram stores fixed duration observations.
type histogram struct {
	// buckets stores bounded duration counts.
	buckets [8]atomic.Uint64
	// count stores all observations.
	count atomic.Uint64
	// nanoseconds stores cumulative duration.
	nanoseconds atomic.Uint64
}

// New creates empty social-group telemetry.
func New() *Metrics { return &Metrics{} }

// Record increments one bounded metric family result.
func (metrics *Metrics) Record(family Family, kind Kind, result Result) {
	if metrics != nil && family < familyCount && kind < kindCount && result < resultCount {
		metrics.families[family][kind][result].Add(1)
	}
}

// Observe records one bounded workflow duration.
func (metrics *Metrics) Observe(kind Timing, duration time.Duration) {
	if metrics == nil || kind >= timingCount {
		return
	}
	thresholds := [...]time.Duration{50 * time.Microsecond, 100 * time.Microsecond, 250 * time.Microsecond, 500 * time.Microsecond, time.Millisecond, 5 * time.Millisecond, 25 * time.Millisecond}
	bucket := len(thresholds)
	for index, threshold := range thresholds {
		if duration <= threshold {
			bucket = index
			break
		}
	}
	metrics.timings[kind].buckets[bucket].Add(1)
	metrics.timings[kind].count.Add(1)
	if duration > 0 {
		metrics.timings[kind].nanoseconds.Add(uint64(duration))
	}
}

// Snapshot returns one consistent-enough lock-free telemetry view.
func (metrics *Metrics) Snapshot() Snapshot {
	if metrics == nil {
		return Snapshot{}
	}
	var snapshot Snapshot
	for family := range snapshot.Families {
		for kind := range snapshot.Families[family] {
			values := &metrics.families[family][kind]
			snapshot.Families[family][kind] = ResultCounters{Success: values[Success].Load(), Rejected: values[Rejected].Load(), Failed: values[Failed].Load(), Unsupported: values[Unsupported].Load()}
		}
	}
	for kind := range snapshot.Timings {
		value := &metrics.timings[kind]
		for bucket := range snapshot.Timings[kind].Buckets {
			snapshot.Timings[kind].Buckets[bucket] = value.buckets[bucket].Load()
		}
		snapshot.Timings[kind].Count = value.count.Load()
		snapshot.Timings[kind].Nanoseconds = value.nanoseconds.Load()
	}
	return snapshot
}
