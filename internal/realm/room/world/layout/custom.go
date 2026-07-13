package layout

import (
	"context"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
)

// RoomManager resolves and persists room-owned layouts.
type RoomManager interface {
	Manager

	// ResolveForRoom resolves a custom layout before its fixed fallback.
	ResolveForRoom(ctx context.Context, roomID int64, modelName string) (Layout, error)
	// SaveCustom creates or replaces a room-owned custom layout.
	SaveCustom(ctx context.Context, params CustomSaveParams) (Layout, error)
	// WithinTransaction runs custom layout work atomically.
	WithinTransaction(ctx context.Context, work TransactionWork) error
}

// CustomManager reads and writes only room-owned custom layouts.
type CustomManager interface {
	// FindCustomByRoomID finds a room-owned layout without falling back to a model.
	FindCustomByRoomID(ctx context.Context, roomID int64) (Layout, bool, error)
	// SaveCustom creates or replaces a room-owned custom layout.
	SaveCustom(ctx context.Context, params CustomSaveParams) (Layout, error)
}

// FindCustomByRoomID finds a room-owned layout without a fixed-model fallback.
func (service *Service) FindCustomByRoomID(ctx context.Context, roomID int64) (Layout, bool, error) {
	custom, ok := service.store.(CustomStore)
	if !ok {
		return Layout{}, false, ErrCustomLayoutsUnsupported
	}
	roomLayout, found, err := custom.FindCustomByRoomID(ctx, roomID)
	if err != nil || !found {
		return Layout{}, found, err
	}
	roomLayout, err = normalizeCustom(roomLayout)
	return roomLayout, err == nil, err
}

// ResolveForRoom resolves room-owned geometry when supported and otherwise uses the fixed layout.
func ResolveForRoom(ctx context.Context, manager Manager, roomID int64, modelName string) (Layout, error) {
	if resolver, ok := manager.(interface {
		ResolveForRoom(context.Context, int64, string) (Layout, error)
	}); ok {
		return resolver.ResolveForRoom(ctx, roomID, modelName)
	}
	roomLayout, found, err := manager.FindByName(ctx, modelName)
	if err != nil {
		return Layout{}, err
	}
	if !found {
		return Layout{}, ErrLayoutNotFound
	}

	return roomLayout, nil
}

// ResolveForRoom resolves a custom layout before its fixed fallback.
func (service *Service) ResolveForRoom(ctx context.Context, roomID int64, modelName string) (Layout, error) {
	if custom, ok := service.store.(CustomStore); ok && roomID > 0 {
		roomLayout, found, err := custom.FindCustomByRoomID(ctx, roomID)
		if err != nil {
			return Layout{}, err
		}
		if found {
			return normalizeCustom(roomLayout)
		}
	}

	roomLayout, found, err := service.FindByName(ctx, modelName)
	if err != nil {
		return Layout{}, err
	}
	if !found {
		return Layout{}, ErrLayoutNotFound
	}

	return roomLayout, nil
}

// SaveCustom creates or replaces a room-owned custom layout.
func (service *Service) SaveCustom(ctx context.Context, params CustomSaveParams) (Layout, error) {
	custom, ok := service.store.(CustomStore)
	if !ok {
		return Layout{}, ErrCustomLayoutsUnsupported
	}
	if params.RoomID <= 0 {
		return Layout{}, ErrInvalidLayoutID
	}

	roomLayout, err := custom.UpsertCustom(ctx, params)
	if err != nil {
		return Layout{}, err
	}

	return normalizeCustom(roomLayout)
}

// WithinTransaction runs custom layout work atomically.
func (service *Service) WithinTransaction(ctx context.Context, work TransactionWork) error {
	custom, ok := service.store.(CustomStore)
	if !ok {
		return ErrCustomLayoutsUnsupported
	}

	return custom.WithinTransaction(ctx, work)
}

// normalizeCustom derives compact fields not stored redundantly in PostgreSQL.
func normalizeCustom(roomLayout Layout) (Layout, error) {
	roomGrid, err := roomLayout.Grid()
	if err != nil {
		return Layout{}, err
	}
	point, ok := grid.NewPoint(roomLayout.DoorX, roomLayout.DoorY)
	if !ok {
		return Layout{}, ErrInvalidLayout
	}
	height, found := roomGrid.HeightAt(point)
	if !found {
		return Layout{}, ErrInvalidLayout
	}
	roomLayout.DoorZ = int(height)
	roomLayout.TileSize = roomGrid.ValidCount()
	roomLayout.Enabled = true

	return roomLayout, nil
}

// roomManagerAssertion verifies Service implements RoomManager.
var roomManagerAssertion RoomManager = (*Service)(nil)
