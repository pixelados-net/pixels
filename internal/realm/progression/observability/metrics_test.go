package observability

import "testing"

// TestMetricsSnapshot verifies counters and gauges are projected independently.
func TestMetricsSnapshot(t *testing.T) {
	metrics := New()
	metrics.RecordTrigger("room.entered")
	metrics.RecordTrigger("room.entered")
	metrics.RecordReward("badge")
	metrics.RecordLevelUps(2)
	metrics.RecordQuestCompleted()
	metrics.AddQueue(3)
	metrics.AddQueue(-1)
	metrics.SetPending(4)
	metrics.SetCacheDefinitions(51)

	snapshot := metrics.Snapshot()
	if snapshot.Triggers["room.entered"] != 2 || snapshot.Rewards["badge"] != 1 {
		t.Fatalf("unexpected bounded counters: %#v", snapshot)
	}
	if snapshot.LevelUps != 2 || snapshot.QuestsCompleted != 1 || snapshot.QueueDepth != 2 || snapshot.PendingFlushes != 4 || snapshot.CacheDefinitions != 51 {
		t.Fatalf("unexpected metric snapshot: %#v", snapshot)
	}
}

// TestNilMetrics verifies optional instrumentation remains safe.
func TestNilMetrics(t *testing.T) {
	var metrics *Metrics
	metrics.RecordTrigger("room.entered")
	metrics.RecordReward("badge")
	metrics.RecordLevelUps(1)
	metrics.RecordQuestCompleted()
	metrics.AddQueue(1)
	metrics.SetPending(1)
	metrics.SetCacheDefinitions(1)
	if snapshot := metrics.Snapshot(); len(snapshot.Triggers) != 0 || len(snapshot.Rewards) != 0 {
		t.Fatalf("unexpected nil snapshot: %#v", snapshot)
	}
}

// BenchmarkRecordTrigger measures the warmed trigger-counter hot path.
func BenchmarkRecordTrigger(b *testing.B) {
	metrics := New()
	metrics.RecordTrigger("room.entered")
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		metrics.RecordTrigger("room.entered")
	}
}
