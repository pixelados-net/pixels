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

// TransactionWork performs custom layout work in one transaction.
type TransactionWork func(context.Context) error

// CustomStore reads and writes room-owned custom layouts.
type CustomStore interface {
	// FindCustomByRoomID finds a room's custom layout.
	FindCustomByRoomID(context.Context, int64) (Layout, bool, error)
	// UpsertCustom creates or replaces a room's custom layout.
	UpsertCustom(context.Context, CustomSaveParams) (Layout, error)
	// WithinTransaction runs work in one shared transaction.
	WithinTransaction(context.Context, TransactionWork) error
}

// CustomSaveParams contains validated custom floor plan values.
type CustomSaveParams struct {
	// RoomID identifies the room owning the floor plan.
	RoomID int64
	// Heightmap stores the compact floor plan.
	Heightmap string
	// DoorX stores the entry tile x coordinate.
	DoorX int
	// DoorY stores the entry tile y coordinate.
	DoorY int
	// DoorZ stores the resolved entry tile height.
	DoorZ int
	// DoorDirection stores the entry direction.
	DoorDirection int
	// WallThickness stores wall rendering thickness.
	WallThickness int
	// FloorThickness stores floor rendering thickness.
	FloorThickness int
	// WallHeight stores fixed wall height or -1 for automatic height.
	WallHeight int
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
