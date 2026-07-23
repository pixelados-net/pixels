// Package repository persists partitioned chat history.
package repository

import (
	"context"
	"time"

	historymodel "github.com/niflaot/pixels/internal/realm/chat/history/model"
)

// Store writes, queries, and maintains chat history partitions.
type Store interface {
	// InsertBatch writes entries through PostgreSQL COPY.
	InsertBatch(context.Context, []historymodel.Entry) error
	// History returns one keyset history page.
	History(context.Context, historymodel.Query) ([]historymodel.Entry, error)
	// EnsurePartitions creates daily partitions through one date.
	EnsurePartitions(context.Context, time.Time, time.Time) error
	// DropBefore drops daily partitions older than a cutoff.
	DropBefore(context.Context, time.Time) error
}

// storeAssertion verifies Repository implements Store.
var storeAssertion Store = (*Repository)(nil)
