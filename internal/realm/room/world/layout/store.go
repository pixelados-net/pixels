package layout

import "context"

// Reader reads persistent room layouts.
type Reader interface {
	// FindByID finds an active room layout by id.
	FindByID(context.Context, int64) (Layout, bool, error)
	// FindByName finds an active room layout by normalized name.
	FindByName(context.Context, string) (Layout, bool, error)
	// List lists active room layouts.
	List(context.Context) ([]Layout, error)
}

// Writer writes persistent room layouts.
type Writer interface {
	// Create creates a room layout record.
	Create(context.Context, CreateRecordParams) (Layout, error)
	// Update updates a room layout record.
	Update(context.Context, UpdateRecordParams) (Layout, bool, error)
}

// Store reads and writes persistent room layouts.
type Store interface {
	Reader
	Writer
}

// CreateRecordParams contains room layout creation data.
type CreateRecordParams struct {
	// Layout contains the room layout values.
	Layout Layout
}

// UpdateRecordParams contains room layout update data.
type UpdateRecordParams struct {
	// ID identifies the room layout record.
	ID int64
	// Layout contains the replacement room layout values.
	Layout Layout
}
