package repository

import roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"

// storeAssertion verifies Repository implements the room persistence contract.
var storeAssertion roomservice.Store = (*Repository)(nil)
