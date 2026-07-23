package history

import (
	"context"
	"testing"
	"time"

	navservice "github.com/niflaot/pixels/internal/realm/navigator/core"
	navrecord "github.com/niflaot/pixels/internal/realm/navigator/record"
	roomentered "github.com/niflaot/pixels/internal/realm/room/access/events/entered"
)

// TestEnqueueIsBounded verifies admission never blocks when telemetry is full.
func TestEnqueueIsBounded(t *testing.T) {
	writer := &Writer{queue: make(chan roomentered.Payload, 1)}
	if !writer.Enqueue(roomentered.Payload{PlayerID: 1, RoomID: 2}) {
		t.Fatal("expected first visit to enqueue")
	}
	if writer.Enqueue(roomentered.Payload{PlayerID: 1, RoomID: 3}) {
		t.Fatal("expected full queue to drop telemetry")
	}
	if writer.Dropped() != 1 {
		t.Fatalf("expected one dropped visit, got %d", writer.Dropped())
	}
}

// historyManager records grouped visit persistence.
type historyManager struct {
	navservice.Manager
	// visits stores the last grouped persistence call.
	visits []navrecord.Visit
}

// RecordVisits stores one grouped visit call.
func (manager *historyManager) RecordVisits(_ context.Context, visits []navrecord.Visit) error {
	manager.visits = append([]navrecord.Visit(nil), visits...)
	return nil
}

// TestPersistBatchCoalescesAndDeduplicates verifies grouped frequency policy.
func TestPersistBatchCoalescesAndDeduplicates(t *testing.T) {
	manager := &historyManager{}
	writer := NewConfigured(manager, nil, 8, time.Minute)
	current := time.Date(2026, time.July, 19, 12, 0, 0, 0, time.UTC)
	writer.now = func() time.Time { return current }
	last := make(map[visitKey]time.Time)
	payloads := []roomentered.Payload{{PlayerID: 1, RoomID: 2}, {PlayerID: 1, RoomID: 2}, {PlayerID: 1, RoomID: 3}}
	writer.persistBatch(context.Background(), payloads, last)
	if len(manager.visits) != 2 || !manager.visits[0].Increment || !manager.visits[1].Increment {
		t.Fatalf("unexpected initial visits %#v", manager.visits)
	}
	current = current.Add(30 * time.Second)
	writer.persistBatch(context.Background(), payloads[:1], last)
	if len(manager.visits) != 1 || manager.visits[0].Increment {
		t.Fatalf("expected duplicate timestamp-only visit, got %#v", manager.visits)
	}
}

// BenchmarkHistoryEnqueue measures the room-entry hot path.
func BenchmarkHistoryEnqueue(b *testing.B) {
	writer := &Writer{queue: make(chan roomentered.Payload, 1)}
	payload := roomentered.Payload{PlayerID: 1, RoomID: 2}
	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		writer.Enqueue(payload)
		<-writer.queue
	}
}
