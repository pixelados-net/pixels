package layout

import (
	"context"
	"fmt"
	"strings"
)

// Service validates and coordinates room layout persistence.
type Service struct {
	// store reads and writes room layout records.
	store Store
}

// SaveParams contains editable room layout input.
type SaveParams struct {
	// Name stores the protocol model name such as model_a.
	Name string

	// TileSize stores the number of walkable/renderable tiles.
	TileSize int

	// Heightmap stores the model heightmap text.
	Heightmap string

	// DoorX stores the door tile x coordinate.
	DoorX int

	// DoorY stores the door tile y coordinate.
	DoorY int

	// DoorZ stores the door tile height.
	DoorZ int

	// DoorDirection stores the door rotation.
	DoorDirection int

	// ClubLevel stores the minimum club level required by the client UI.
	ClubLevel int

	// Enabled reports whether the layout can be used for new rooms.
	Enabled bool
}

// Manager manages persistent room layouts.
type Manager interface {
	// Create creates a room layout.
	Create(ctx context.Context, params SaveParams) (Layout, error)

	// Update updates a room layout.
	Update(ctx context.Context, id int64, params SaveParams) (Layout, error)

	// FindByID finds an active room layout by id.
	FindByID(ctx context.Context, id int64) (Layout, bool, error)

	// FindByName finds an active room layout by normalized name.
	FindByName(ctx context.Context, name string) (Layout, bool, error)

	// List lists active room layouts.
	List(ctx context.Context) ([]Layout, error)

	// Catalog loads active room layouts into an in-memory catalog.
	Catalog(ctx context.Context) (*Catalog, error)
}

// NewService creates a room layout service.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// Create creates a room layout.
func (service *Service) Create(ctx context.Context, params SaveParams) (Layout, error) {
	roomLayout, err := layoutFromParams(params)
	if err != nil {
		return Layout{}, err
	}

	return service.store.Create(ctx, CreateRecordParams{Layout: roomLayout})
}

// Update updates a room layout.
func (service *Service) Update(ctx context.Context, id int64, params SaveParams) (Layout, error) {
	if id <= 0 {
		return Layout{}, ErrInvalidLayoutID
	}

	roomLayout, err := layoutFromParams(params)
	if err != nil {
		return Layout{}, err
	}

	updated, found, err := service.store.Update(ctx, UpdateRecordParams{ID: id, Layout: roomLayout})
	if err != nil {
		return Layout{}, err
	}

	if !found {
		return Layout{}, ErrLayoutNotFound
	}

	return updated, nil
}

// FindByID finds an active room layout by id.
func (service *Service) FindByID(ctx context.Context, id int64) (Layout, bool, error) {
	if id <= 0 {
		return Layout{}, false, ErrInvalidLayoutID
	}

	return service.store.FindByID(ctx, id)
}

// FindByName finds an active room layout by normalized name.
func (service *Service) FindByName(ctx context.Context, name string) (Layout, bool, error) {
	name = NormalizeName(name)
	if name == "" {
		return Layout{}, false, ErrInvalidLayout
	}

	return service.store.FindByName(ctx, name)
}

// List lists active room layouts.
func (service *Service) List(ctx context.Context) ([]Layout, error) {
	return service.store.List(ctx)
}

// Catalog loads active room layouts into an in-memory catalog.
func (service *Service) Catalog(ctx context.Context) (*Catalog, error) {
	layouts, err := service.store.List(ctx)
	if err != nil {
		return nil, err
	}

	catalog, err := NewCatalog(layouts)
	if err != nil {
		return nil, fmt.Errorf("build room layout catalog: %w", err)
	}

	return catalog, nil
}

// layoutFromParams maps editable input to a validated room layout.
func layoutFromParams(params SaveParams) (Layout, error) {
	roomLayout := Layout{
		Name:          NormalizeName(params.Name),
		TileSize:      params.TileSize,
		Heightmap:     strings.TrimSpace(params.Heightmap),
		DoorX:         params.DoorX,
		DoorY:         params.DoorY,
		DoorZ:         params.DoorZ,
		DoorDirection: params.DoorDirection,
		ClubLevel:     params.ClubLevel,
		Enabled:       params.Enabled,
	}
	if !roomLayout.Valid() {
		return Layout{}, ErrInvalidLayout
	}

	return roomLayout, nil
}

// managerAssertion verifies Service implements Manager.
var managerAssertion Manager = (*Service)(nil)
