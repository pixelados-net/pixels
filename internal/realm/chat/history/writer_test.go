package history

import (
	"context"
	"sync"
	"testing"
	"time"

	chatconfig "github.com/niflaot/pixels/internal/realm/chat/config"
	shoutedevent "github.com/niflaot/pixels/internal/realm/chat/events/shouted"
	talkedevent "github.com/niflaot/pixels/internal/realm/chat/events/talked"
	whisperedevent "github.com/niflaot/pixels/internal/realm/chat/events/whispered"
	historymodel "github.com/niflaot/pixels/internal/realm/chat/history/model"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx/fxtest"
	"go.uber.org/zap"
)

// storeForTest records history batches.
type storeForTest struct {
	// mutex protects batches.
	mutex sync.Mutex
	// batches stores copied history batches.
	batches [][]historymodel.Entry
	// flushed signals one completed batch.
	flushed chan struct{}
	// history stores query fixture rows.
	history []historymodel.Entry
	// maintained counts partition maintenance calls.
	maintained int
}

// InsertBatch records a detached batch.
func (store *storeForTest) InsertBatch(_ context.Context, entries []historymodel.Entry) error {
	store.mutex.Lock()
	store.batches = append(store.batches, append([]historymodel.Entry(nil), entries...))
	store.mutex.Unlock()
	select {
	case store.flushed <- struct{}{}:
	default:
	}
	return nil
}

// History returns no fixture entries.
func (store *storeForTest) History(context.Context, historymodel.Query) ([]historymodel.Entry, error) {
	return store.history, nil
}

// EnsurePartitions accepts maintenance.
func (store *storeForTest) EnsurePartitions(context.Context, time.Time, time.Time) error {
	store.maintained++
	return nil
}

// DropBefore accepts maintenance.
func (store *storeForTest) DropBefore(context.Context, time.Time) error {
	store.maintained++
	return nil
}

// TestWriterFlushesFullBatch verifies capacity-triggered asynchronous writes.
func TestWriterFlushesFullBatch(t *testing.T) {
	store := &storeForTest{flushed: make(chan struct{}, 1)}
	writer := NewWriter(chatconfig.Config{HistoryBatchSize: 2, HistoryQueueSize: 4, HistoryFlushInterval: time.Hour}, store, zap.NewNop())
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { writer.run(ctx); close(done) }()
	writer.Enqueue(historymodel.Entry{RoomID: 1, Message: "one"})
	writer.Enqueue(historymodel.Entry{RoomID: 1, Message: "two"})
	select {
	case <-store.flushed:
	case <-time.After(time.Second):
		t.Fatal("history batch did not flush")
	}
	cancel()
	<-done
	if len(store.batches) == 0 || len(store.batches[0]) != 2 {
		t.Fatalf("unexpected batches %#v", store.batches)
	}
}

// TestRegisterPersistsConfiguredEventFamilies verifies bus and lifecycle wiring.
func TestRegisterPersistsConfiguredEventFamilies(t *testing.T) {
	store := &storeForTest{flushed: make(chan struct{}, 3)}
	config := chatconfig.Config{HistoryBatchSize: 1, HistoryQueueSize: 4, HistoryFlushInterval: time.Hour, LogWhispers: true}
	writer := NewWriter(config, store, zap.NewNop())
	local := bus.New()
	lifecycle := fxtest.NewLifecycle(t)
	if err := Register(lifecycle, local, writer, store, config, zap.NewNop()); err != nil {
		t.Fatalf("register: %v", err)
	}
	lifecycle.RequireStart()
	events := []bus.Event{
		{Name: talkedevent.Name, Payload: talkedevent.Payload{RoomID: 1, PlayerID: 2, Message: "talk", CreatedAt: time.Now()}},
		{Name: shoutedevent.Name, Payload: shoutedevent.Payload{RoomID: 1, PlayerID: 2, Message: "shout", CreatedAt: time.Now()}},
		{Name: whisperedevent.Name, Payload: whisperedevent.Payload{RoomID: 1, PlayerID: 2, TargetPlayerID: 3, Message: "whisper", CreatedAt: time.Now()}},
	}
	for _, event := range events {
		if err := local.Publish(context.Background(), event); err != nil {
			t.Fatalf("publish: %v", err)
		}
	}
	for range events {
		select {
		case <-store.flushed:
		case <-time.After(time.Second):
			t.Fatal("event did not flush")
		}
	}
	lifecycle.RequireStop()
	if store.maintained != 2 || len(store.batches) < 3 {
		t.Fatalf("maintained=%d batches=%d", store.maintained, len(store.batches))
	}
}

// TestServiceReturnsHistory verifies bounded query delegation.
func TestServiceReturnsHistory(t *testing.T) {
	store := &storeForTest{history: []historymodel.Entry{{ID: 1}}}
	items, err := NewService(store).History(context.Background(), historymodel.Query{Limit: 500})
	if err != nil || len(items) != 1 {
		t.Fatalf("items=%#v err=%v", items, err)
	}
}

// TestWriterNeverBlocksWhenFull verifies overload drops instead of backpressure.
func TestWriterNeverBlocksWhenFull(t *testing.T) {
	writer := NewWriter(chatconfig.Config{HistoryBatchSize: 1, HistoryQueueSize: 1}, &storeForTest{}, zap.NewNop())
	if !writer.Enqueue(historymodel.Entry{Message: "accepted"}) {
		t.Fatal("expected first entry accepted")
	}
	if writer.Enqueue(historymodel.Entry{Message: "dropped"}) || writer.Dropped() != 1 {
		t.Fatalf("expected one dropped entry, got %d", writer.Dropped())
	}
}

// BenchmarkEnqueue measures the bounded history publication path.
func BenchmarkEnqueue(b *testing.B) {
	writer := NewWriter(chatconfig.Config{HistoryBatchSize: 1, HistoryQueueSize: 1}, &storeForTest{}, zap.NewNop())
	entry := historymodel.Entry{RoomID: 1, PlayerID: 2, Kind: "talk", Message: "hello", CreatedAt: time.Now()}
	b.ReportAllocs()
	for range b.N {
		writer.Enqueue(entry)
	}
}
