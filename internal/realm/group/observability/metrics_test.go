package observability

import (
	"testing"
	"time"
)

// TestMetricsRemainBoundedAndSnapshot verifies fixed counters and histograms.
func TestMetricsRemainBoundedAndSnapshot(t *testing.T) {
	metrics := New()
	metrics.Record(Membership, KindJoin, Success)
	metrics.Record(Membership, KindJoin, Unsupported)
	metrics.Record(Family(99), KindDefault, Failed)
	metrics.Observe(ForumQuery, time.Millisecond)
	snapshot := metrics.Snapshot()
	if snapshot.Families[Membership][KindJoin].Success != 1 || snapshot.Families[Membership][KindJoin].Unsupported != 1 || snapshot.Timings[ForumQuery].Count != 1 {
		t.Fatalf("snapshot=%#v", snapshot)
	}
}
