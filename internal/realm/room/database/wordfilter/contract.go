// Package repository persists room-specific filtered words.
package repository

import roomwordfilter "github.com/niflaot/pixels/internal/realm/room/control/wordfilter"

// storeAssertion verifies Repository implements the word filter store contract.
var storeAssertion roomwordfilter.Store = (*Repository)(nil)
