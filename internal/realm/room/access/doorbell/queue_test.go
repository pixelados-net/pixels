package doorbell

import (
	"sync"
	"testing"
	"time"

	netconn "github.com/niflaot/pixels/networking/connection"
)

// TestQueueRefreshResolveAndSweep verifies request lifecycle without timers.
func TestQueueRefreshResolveAndSweep(t *testing.T) {
	now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
	var queue Queue
	entry := queueEntry(7, "Demo", now)
	if !queue.Request(entry) || !queue.Request(queueEntry(7, "Demo", now.Add(time.Minute))) || queue.Len() != 1 {
		t.Fatal("expected one refreshed request")
	}
	if expired := queue.Sweep(now.Add(5*time.Minute), 5*time.Minute); len(expired) != 0 {
		t.Fatalf("expected refreshed request to remain, got %#v", expired)
	}
	resolved, found := queue.Resolve("demo")
	if !found || resolved.PlayerID != 7 || queue.Len() != 0 {
		t.Fatalf("unexpected resolution %#v found=%v", resolved, found)
	}
	queue.Request(queueEntry(8, "Other", now))
	expired := queue.Sweep(now.Add(5*time.Minute), 5*time.Minute)
	if len(expired) != 1 || expired[0].Reason != ExpiredTimeout {
		t.Fatalf("unexpected expiration %#v", expired)
	}
}

// TestQueueValidationAndDrain verifies invalid, missing, and full-drain paths.
func TestQueueValidationAndDrain(t *testing.T) {
	var queue Queue
	if queue.Request(Entry{}) {
		t.Fatal("expected invalid request rejection")
	}
	if _, found := queue.Resolve("missing"); found {
		t.Fatal("expected missing resolution")
	}
	if expired := queue.Sweep(time.Now(), 0); expired != nil {
		t.Fatalf("expected disabled sweep, got %#v", expired)
	}
	if drained := queue.Drain(ExpiredRoomClosed); drained != nil {
		t.Fatalf("expected empty drain, got %#v", drained)
	}
	now := time.Now()
	queue.Request(queueEntry(7, "Demo", now))
	queue.Request(queueEntry(8, "Other", now))
	drained := queue.Drain(ExpiredRoomClosed)
	if len(drained) != 2 || queue.Len() != 0 {
		t.Fatalf("unexpected drained entries %#v", drained)
	}
}

// TestQueueConcurrentRefresh verifies queue synchronization under concurrent requests.
func TestQueueConcurrentRefresh(t *testing.T) {
	var queue Queue
	now := time.Now()
	var group sync.WaitGroup
	for index := 0; index < 32; index++ {
		group.Add(1)
		go func() {
			defer group.Done()
			queue.Request(queueEntry(7, "Demo", now))
		}()
	}
	group.Wait()
	if queue.Len() != 1 {
		t.Fatalf("expected one request, got %d", queue.Len())
	}
}

// BenchmarkQueueRefresh measures the allocation-free existing-entry path.
func BenchmarkQueueRefresh(b *testing.B) {
	var queue Queue
	entry := queueEntry(7, "Demo", time.Now())
	queue.Request(entry)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		queue.Request(entry)
	}
}

// queueEntry creates one waiting request fixture.
func queueEntry(playerID int64, username string, requestedAt time.Time) Entry {
	return Entry{PlayerID: playerID, Username: username, Handler: netconn.Context{ConnectionID: "conn", ConnectionKind: "websocket"}, RequestedAt: requestedAt}
}
