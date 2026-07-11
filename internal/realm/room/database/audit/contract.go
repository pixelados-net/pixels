// Package repository persists room audit history.
package repository

import roomaudit "github.com/niflaot/pixels/internal/realm/room/control/audit"

// storeAssertion verifies Repository implements the room audit store contract.
var storeAssertion roomaudit.Store = (*Repository)(nil)
