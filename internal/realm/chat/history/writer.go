package history

import (
	"context"
	"sync/atomic"
	"time"

	chatconfig "github.com/niflaot/pixels/internal/realm/chat/config"
	historymodel "github.com/niflaot/pixels/internal/realm/chat/history/model"
	historyrepo "github.com/niflaot/pixels/internal/realm/chat/history/repository"
	"go.uber.org/zap"
)

// Writer buffers history entries without blocking live chat delivery.
type Writer struct {
	// config stores normalized batching settings.
	config chatconfig.Config
	// store persists history batches.
	store historyrepo.Store
	// entries stores the bounded non-blocking queue.
	entries chan historymodel.Entry
	// dropped counts entries discarded under sustained overload.
	dropped atomic.Uint64
	// log records persistence failures with batch context.
	log *zap.Logger
}

// NewWriter creates a bounded chat history writer.
func NewWriter(config chatconfig.Config, store historyrepo.Store, log *zap.Logger) *Writer {
	config = config.Normalize()
	if log == nil {
		log = zap.NewNop()
	}
	return &Writer{config: config, store: store, entries: make(chan historymodel.Entry, config.HistoryQueueSize), log: log}
}

// Enqueue buffers one entry and reports whether it was accepted.
func (writer *Writer) Enqueue(entry historymodel.Entry) bool {
	select {
	case writer.entries <- entry:
		return true
	default:
		writer.dropped.Add(1)
		return false
	}
}

// Dropped returns the overload discard count.
func (writer *Writer) Dropped() uint64 { return writer.dropped.Load() }

// run flushes batches on capacity, interval, and shutdown.
func (writer *Writer) run(ctx context.Context) {
	ticker := time.NewTicker(writer.config.HistoryFlushInterval)
	defer ticker.Stop()
	batch := make([]historymodel.Entry, 0, writer.config.HistoryBatchSize)
	for {
		select {
		case entry := <-writer.entries:
			batch = append(batch, entry)
			if len(batch) >= writer.config.HistoryBatchSize {
				batch = writer.flush(ctx, batch)
			}
		case <-ticker.C:
			batch = writer.flush(ctx, batch)
		case <-ctx.Done():
			writer.drain(batch)
			return
		}
	}
}

// flush writes one batch and reuses its backing storage.
func (writer *Writer) flush(ctx context.Context, batch []historymodel.Entry) []historymodel.Entry {
	if len(batch) == 0 {
		return batch
	}
	if err := writer.store.InsertBatch(ctx, batch); err != nil {
		writer.dropped.Add(uint64(len(batch)))
		writer.log.Error("chat history batch write failed", zap.Error(err), zap.Int("entries", len(batch)))
	}

	return batch[:0]
}

// drain writes every queued entry in bounded shutdown batches.
func (writer *Writer) drain(batch []historymodel.Entry) {
	ctx := context.Background()
	for len(writer.entries) > 0 {
		for len(writer.entries) > 0 && len(batch) < writer.config.HistoryBatchSize {
			batch = append(batch, <-writer.entries)
		}
		batch = writer.flush(ctx, batch)
	}
	writer.flush(ctx, batch)
}
