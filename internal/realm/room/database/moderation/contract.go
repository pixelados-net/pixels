package repository

import roommoderation "github.com/niflaot/pixels/internal/realm/room/control/moderation"

// storeAssertion verifies Repository implements the moderation store contract.
var storeAssertion roommoderation.Store = (*Repository)(nil)
