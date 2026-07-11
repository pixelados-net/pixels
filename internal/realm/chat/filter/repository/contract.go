// Package repository persists global chat filter words.
package repository

import "context"

// Store reads and mutates the global filter dictionary.
type Store interface {
	// List returns normalized filter words.
	List(context.Context) ([]string, error)
	// Add creates a filter word when absent.
	Add(context.Context, string) error
	// Remove deletes a filter word when present.
	Remove(context.Context, string) error
}

// storeAssertion verifies Repository implements Store.
var storeAssertion Store = (*Repository)(nil)
