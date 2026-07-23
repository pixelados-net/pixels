package record

import (
	"time"

	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// SavedSearch stores a player navigator saved search.
type SavedSearch struct {
	// Identity contains the durable record identifier.
	sharedmodel.Identity

	// PlayerID identifies the player.
	PlayerID int64

	// Code stores the navigator context or result code.
	Code string

	// Filter stores the saved search query.
	Filter string

	// Localization stores an optional localization key.
	Localization string

	// CreatedAt is the time the search was saved.
	CreatedAt time.Time
}
