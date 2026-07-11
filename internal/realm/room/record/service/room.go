package service

import (
	"context"
	"fmt"
	"strings"

	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
)

// CreateParams contains room creation input.
type CreateParams struct {
	// OwnerPlayerID identifies the player that owns the room.
	OwnerPlayerID int64

	// OwnerName stores an owner name snapshot for navigator listings.
	OwnerName string

	// Name is the visible room name.
	Name string

	// Description is the visible room description.
	Description string

	// ModelName is the room layout model name.
	ModelName string

	// MaxUsers stores the maximum user count.
	MaxUsers int

	// CategoryID optionally identifies the navigator category.
	CategoryID *int64

	// TradeMode describes trading behavior.
	TradeMode roommodel.TradeMode

	// Tags stores room tags.
	Tags []string
}

// Create creates a room and its tags.
func (service *Service) Create(ctx context.Context, params CreateParams) (roommodel.Room, error) {
	params = normalizeCreate(params)
	if err := validateCreate(params); err != nil {
		return roommodel.Room{}, err
	}
	if err := service.validateCategory(ctx, params.CategoryID, false); err != nil {
		return roommodel.Room{}, err
	}
	tags := normalizeTags(params.Tags)
	if err := service.validateContent(ctx, roommodel.Room{Name: params.Name, Description: params.Description}, tags); err != nil {
		return roommodel.Room{}, err
	}

	roomLayout, found, err := service.layouts.FindByName(ctx, params.ModelName)
	if err != nil {
		return roommodel.Room{}, fmt.Errorf("find room layout: %w", err)
	}
	if !found || !roomLayout.Enabled {
		return roommodel.Room{}, ErrLayoutNotAvailable
	}

	room, err := service.store.CreateRoom(ctx, createRoomParams(params))
	if err != nil {
		return roommodel.Room{}, err
	}

	if err := service.store.ReplaceRoomTags(ctx, room.ID, tags); err != nil {
		return roommodel.Room{}, fmt.Errorf("replace room tags: %w", err)
	}

	return room, nil
}

// FindByID finds a room by id.
func (service *Service) FindByID(ctx context.Context, id int64) (roommodel.Room, bool, error) {
	if id <= 0 {
		return roommodel.Room{}, false, ErrInvalidRoomID
	}

	return service.store.FindRoomByID(ctx, id)
}

// ListByOwner lists rooms owned by a player.
func (service *Service) ListByOwner(ctx context.Context, ownerPlayerID int64) ([]roommodel.Room, error) {
	if ownerPlayerID <= 0 {
		return nil, ErrInvalidOwner
	}

	return service.store.ListRoomsByOwner(ctx, ownerPlayerID)
}

// ListPopular lists popular rooms.
func (service *Service) ListPopular(ctx context.Context, limit int) ([]roommodel.Room, error) {
	return service.store.ListPopularRooms(ctx, normalizeLimit(limit))
}

// ListHighestScore lists highest scoring rooms.
func (service *Service) ListHighestScore(ctx context.Context, limit int) ([]roommodel.Room, error) {
	return service.store.ListHighestScoreRooms(ctx, normalizeLimit(limit))
}

// Search searches public room navigator fields.
func (service *Service) Search(ctx context.Context, query string, limit int) ([]roommodel.Room, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return service.ListPopular(ctx, limit)
	}

	return service.store.SearchRooms(ctx, query, normalizeLimit(limit))
}

// ListTags lists normalized room tags.
func (service *Service) ListTags(ctx context.Context, roomID int64) ([]roommodel.Tag, error) {
	if roomID <= 0 {
		return nil, ErrInvalidRoomID
	}

	return service.store.ListRoomTags(ctx, roomID)
}

// SoftDelete soft deletes a room.
func (service *Service) SoftDelete(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrInvalidRoomID
	}

	deleted, err := service.store.SoftDeleteRoom(ctx, id)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrRoomNotFound
	}

	return nil
}

// ListCategories lists room categories.
func (service *Service) ListCategories(ctx context.Context) ([]roommodel.Category, error) {
	return service.store.ListCategories(ctx)
}

// normalizeCreate normalizes room creation input.
func normalizeCreate(params CreateParams) CreateParams {
	params.OwnerName = strings.TrimSpace(params.OwnerName)
	params.Name = strings.TrimSpace(params.Name)
	params.Description = strings.TrimSpace(params.Description)
	params.ModelName = strings.TrimSpace(params.ModelName)
	if params.MaxUsers == 0 {
		params.MaxUsers = DefaultMaxUsers
	}

	return params
}

// createRoomParams maps service input to repository input.
func createRoomParams(params CreateParams) CreateRecordParams {
	return CreateRecordParams{
		OwnerPlayerID: params.OwnerPlayerID,
		OwnerName:     params.OwnerName,
		Name:          params.Name,
		Description:   params.Description,
		ModelName:     params.ModelName,
		DoorMode:      roommodel.DoorModeOpen,
		MaxUsers:      params.MaxUsers,
		CategoryID:    params.CategoryID,
		TradeMode:     params.TradeMode,
	}
}

// normalizeLimit normalizes list limits.
func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 50
	}
	if limit > 100 {
		return 100
	}

	return limit
}
