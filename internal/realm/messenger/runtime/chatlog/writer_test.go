package chatlog

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"go.uber.org/zap"
)

// TestWriterPersistsQueuedMessage verifies asynchronous draining.
func TestWriterPersistsQueuedMessage(t *testing.T) {
	store := &fakeMessageStore{written: make(chan struct{}, 1)}
	writer := New(Config{Enabled: true}, store, zap.NewNop())
	ctx, cancel := context.WithCancel(context.Background())
	writer.Start(ctx)
	if !writer.Enqueue(1, 2, "hello") {
		t.Fatal("expected message to be queued")
	}
	select {
	case <-store.written:
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for persistence")
	}
	cancel()
	writer.Wait()
	if store.count.Load() != 1 {
		t.Fatalf("expected one write, got %d", store.count.Load())
	}
}

// BenchmarkEnqueue measures the nonblocking private-chat hot path.
func BenchmarkEnqueue(b *testing.B) {
	writer := New(Config{}, &fakeMessageStore{}, zap.NewNop())
	b.ReportAllocs()
	for b.Loop() {
		writer.Enqueue(1, 2, "hello")
	}
}

// fakeMessageStore counts persisted private messages.
type fakeMessageStore struct {
	// count stores completed writes.
	count atomic.Int64
	// written signals a completed write.
	written chan struct{}
}

// LogPrivateMessage records one write.
func (store *fakeMessageStore) LogPrivateMessage(context.Context, int64, int64, string) error {
	store.count.Add(1)
	if store.written != nil {
		store.written <- struct{}{}
	}
	return nil
}
