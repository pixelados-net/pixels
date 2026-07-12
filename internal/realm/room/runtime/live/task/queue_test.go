package task

import (
	"testing"
	"time"
)

// TestQueueRunsDueAndReplaces verifies deadlines, replacement, and cleanup.
func TestQueueRunsDueAndReplaces(t *testing.T) {
	now := time.Unix(100, 0)
	var queue Queue
	value := 0
	queue.Schedule(now.Add(time.Second), func(time.Time) { value++ })
	queue.Replace(7, now.Add(time.Second), func(time.Time) { value += 10 })
	queue.Replace(7, now.Add(2*time.Second), func(time.Time) { value += 100 })
	if due := queue.Due(now.Add(time.Second)); len(due) != 1 {
		t.Fatalf("expected one due task, got %d", len(due))
	} else {
		due[0](now.Add(time.Second))
	}
	if value != 1 || queue.Len() != 1 {
		t.Fatalf("unexpected state value=%d pending=%d", value, queue.Len())
	}
	for _, run := range queue.Due(now.Add(2 * time.Second)) {
		run(now.Add(2 * time.Second))
	}
	if value != 101 || queue.Len() != 0 {
		t.Fatalf("unexpected final state value=%d pending=%d", value, queue.Len())
	}
	queue.Schedule(now, func(time.Time) {})
	queue.Clear()
	if queue.Len() != 0 {
		t.Fatal("expected queue cleanup")
	}
}

// BenchmarkDue measures a realistic room task sweep.
func BenchmarkDue(b *testing.B) {
	now := time.Unix(100, 0)
	b.ReportAllocs()
	for b.Loop() {
		var queue Queue
		for index := 0; index < 20; index++ {
			queue.Schedule(now, func(time.Time) {})
		}
		_ = queue.Due(now)
	}
}
