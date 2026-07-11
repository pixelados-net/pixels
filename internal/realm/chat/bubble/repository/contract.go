// Package repository persists chat bubble unlock thresholds.
package repository

import "context"

// Unlock describes one configured bubble threshold.
type Unlock struct {
	// BubbleID identifies the Nitro bubble style.
	BubbleID int32 `json:"bubbleId"`
	// MinWeight stores the minimum primary-group weight.
	MinWeight int32 `json:"minWeight"`
}

// Store reads and mutates chat bubble thresholds.
type Store interface {
	// List returns configured thresholds.
	List(context.Context) ([]Unlock, error)
	// MinWeight returns one threshold and whether it exists.
	MinWeight(context.Context, int32) (int32, bool, error)
	// Set creates or replaces one threshold.
	Set(context.Context, int32, int32) error
	// Delete removes one threshold.
	Delete(context.Context, int32) error
}

// storeAssertion verifies Repository implements Store.
var storeAssertion Store = (*Repository)(nil)
