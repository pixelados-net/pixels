package model

import "time"

// Timestamps tracks durable record creation and update times.
type Timestamps struct {
	// CreatedAt is the time the record was created.
	CreatedAt time.Time

	// UpdatedAt is the time the record was last updated.
	UpdatedAt time.Time
}

// Touch updates the record update time.
func (timestamps *Timestamps) Touch(now time.Time) {
	timestamps.UpdatedAt = now
}
