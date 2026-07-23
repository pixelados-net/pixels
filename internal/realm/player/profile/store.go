package profile

import (
	"context"
	"time"
)

// Store persists public tags and serialized daily respect grants.
type Store interface {
	// Tags returns ordered public tags.
	Tags(context.Context, int64) ([]string, error)
	// ReplaceTags atomically replaces ordered public tags.
	ReplaceTags(context.Context, int64, []string) error
	// RespectState returns current totals and daily allowances.
	RespectState(context.Context, int64, time.Time, int, int) (RespectState, error)
	// GrantRespect serializes and applies one daily user respect.
	GrantRespect(context.Context, int64, int64, time.Time, int, bool) (RespectResult, error)
}
