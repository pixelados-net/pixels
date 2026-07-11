// Package repository persists room-specific filtered words.
package repository

import "context"

// Store reads and mutates room word filters.
type Store interface {
	// List lists normalized words for a room.
	List(context.Context, int64) ([]string, error)
	// Add inserts a normalized room word.
	Add(context.Context, int64, string) error
	// Remove deletes a normalized room word.
	Remove(context.Context, int64, string) error
}
