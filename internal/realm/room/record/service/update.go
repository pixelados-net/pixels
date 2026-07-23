package service

import (
	"context"
	"strings"

	roomentry "github.com/niflaot/pixels/internal/realm/room/access/entry"
	roomstaffpicked "github.com/niflaot/pixels/internal/realm/room/record/events/staffpicked"
	roomupdated "github.com/niflaot/pixels/internal/realm/room/record/events/updated"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	"github.com/niflaot/pixels/pkg/bus"
	"golang.org/x/crypto/bcrypt"
)

// ProfanityChecker detects prohibited user-facing content.
type ProfanityChecker interface {
	// Contains reports whether text contains prohibited content.
	Contains(context.Context, string) (bool, error)
}

// UpdateParams contains optional editable room settings.
type UpdateParams struct {
	// Name replaces the visible room name.
	Name *string
	// Description replaces the visible room description.
	Description *string
	// CategoryID replaces or clears the navigator category.
	CategoryID **int64
	// Tags replaces the complete normalized tag set.
	Tags *[]string
	// MaxUsers replaces room capacity.
	MaxUsers *int
	// DoorMode replaces room access mode.
	DoorMode *roommodel.DoorMode
	// Password contains a new plaintext password and is never persisted directly.
	Password *string
	// TradeMode replaces room trading behavior.
	TradeMode *roommodel.TradeMode
	// RollerSpeed replaces owner-loop cycles between roller steps.
	RollerSpeed *int
	// AllowWalkthrough replaces unit walkthrough behavior.
	AllowWalkthrough *bool
	// AllowPets replaces pet admission behavior.
	AllowPets *bool
	// AllowPetsEat replaces pet food behavior.
	AllowPetsEat *bool
	// HideWalls replaces wall visibility.
	HideWalls *bool
	// HideWired replaces WIRED configuration-box visibility.
	HideWired *bool
	// WallThickness replaces wall thickness.
	WallThickness *int
	// FloorThickness replaces floor thickness.
	FloorThickness *int
	// ChatMode replaces room chat mode.
	ChatMode *int16
	// ChatWeight replaces room chat bubble weight.
	ChatWeight *int16
	// ChatSpeed replaces room chat speed.
	ChatSpeed *int16
	// ChatDistance replaces room chat distance.
	ChatDistance *int16
	// ChatProtection replaces flood protection.
	ChatProtection *int16
	// ModerationMute replaces mute policy.
	ModerationMute *roommodel.ModerationPolicy
	// ModerationKick replaces kick policy.
	ModerationKick *roommodel.ModerationPolicy
	// ModerationBan replaces ban policy.
	ModerationBan *roommodel.ModerationPolicy
	// StaffPicked replaces official Navigator selection for administrative callers.
	StaffPicked *bool
	// AllowReservedTags permits staff-only tag prefixes.
	AllowReservedTags bool
}

// Update applies a partial settings mutation with optimistic locking.
func (service *Service) Update(ctx context.Context, roomID int64, expectedVersion int64, params UpdateParams) (roommodel.Room, error) {
	current, found, err := service.FindByID(ctx, roomID)
	if err != nil {
		return roommodel.Room{}, err
	}
	if !found {
		return roommodel.Room{}, ErrRoomNotFound
	}
	updated, tags, err := service.mergeUpdate(ctx, current, params)
	if err != nil {
		return roommodel.Room{}, err
	}
	if params.Tags == nil {
		existing, listErr := service.ListTags(ctx, roomID)
		if listErr != nil {
			return roommodel.Room{}, listErr
		}
		tags = make([]string, 0, len(existing))
		for _, tag := range existing {
			tags = append(tags, tag.Value)
		}
	}
	updated, found, err = service.store.UpdateRoom(ctx, UpdateRecordParams{Room: updated, ExpectedVersion: expectedVersion}, tags)
	if err != nil {
		return roommodel.Room{}, err
	}
	if !found {
		return roommodel.Room{}, ErrVersionConflict
	}
	if service.events != nil {
		_ = service.events.Publish(ctx, bus.Event{Name: roomupdated.Name, Payload: roomupdated.Payload{RoomID: updated.ID}})
		if params.StaffPicked != nil && *params.StaffPicked && !current.StaffPicked {
			_ = service.events.Publish(ctx, bus.Event{Name: roomstaffpicked.Name, Payload: roomstaffpicked.Payload{RoomID: updated.ID, OwnerPlayerID: updated.OwnerPlayerID}})
		}
	}

	return updated, nil
}

// mergeUpdate applies and validates optional fields in memory.
func (service *Service) mergeUpdate(ctx context.Context, room roommodel.Room, params UpdateParams) (roommodel.Room, []string, error) {
	applyUpdate(&room, params)
	tags := []string(nil)
	if params.Tags != nil {
		tags = normalizeTags(*params.Tags)
	}
	if err := validateUpdate(room, params, tags); err != nil {
		return roommodel.Room{}, nil, err
	}
	if err := service.validateCategory(ctx, room.CategoryID, params.AllowReservedTags); err != nil {
		return roommodel.Room{}, nil, err
	}
	if err := service.validateContent(ctx, room, tags); err != nil {
		return roommodel.Room{}, nil, err
	}
	if params.Password != nil && strings.TrimSpace(*params.Password) != "" {
		hash, err := roomentry.HashPassword(*params.Password, bcrypt.DefaultCost)
		if err != nil {
			return roommodel.Room{}, nil, err
		}
		room.PasswordHash = &hash
	}
	if room.DoorMode == roommodel.DoorModePassword && room.PasswordHash == nil {
		return roommodel.Room{}, nil, ErrPasswordRequired
	}

	return room, tags, nil
}

// validateCategory requires one selectable room category when configured.
func (service *Service) validateCategory(ctx context.Context, categoryID *int64, allowStaff bool) error {
	if categoryID == nil {
		return nil
	}
	categories, err := service.ListCategories(ctx)
	if err != nil {
		return err
	}
	for _, category := range categories {
		if category.ID != *categoryID {
			continue
		}
		if !category.Visible || (category.StaffOnly && !allowStaff) {
			return ErrInvalidCategory
		}

		return nil
	}

	return ErrInvalidCategory
}

// validateContent checks optional global prohibited-content behavior.
func (service *Service) validateContent(ctx context.Context, room roommodel.Room, tags []string) error {
	if service.profanity == nil {
		return nil
	}
	blocked, err := service.profanity.Contains(ctx, room.Name)
	if err != nil {
		return err
	}
	if blocked {
		return ErrProhibitedName
	}
	blocked, err = service.profanity.Contains(ctx, room.Description)
	if err != nil {
		return err
	}
	if blocked {
		return ErrProhibitedDescription
	}
	for _, value := range tags {
		blocked, err := service.profanity.Contains(ctx, value)
		if err != nil {
			return err
		}
		if blocked {
			return ErrProhibitedTag
		}
	}

	return nil
}
