package model

import "time"

// SoftDelete tracks optional durable record deletion.
type SoftDelete struct {
	// DeletedAt is the time the record was soft deleted.
	DeletedAt *time.Time
}

// Active reports whether the record is not soft deleted.
func (softDelete SoftDelete) Active() bool {
	return softDelete.DeletedAt == nil
}

// Deleted reports whether the record is soft deleted.
func (softDelete SoftDelete) Deleted() bool {
	return softDelete.DeletedAt != nil
}
