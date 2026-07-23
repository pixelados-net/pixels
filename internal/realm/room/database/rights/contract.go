package repository

import roomrights "github.com/niflaot/pixels/internal/realm/room/control/rights"

// storeAssertion verifies Repository implements the room rights store contract.
var storeAssertion roomrights.Store = (*Repository)(nil)
