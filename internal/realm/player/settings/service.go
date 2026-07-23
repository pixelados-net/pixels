package settings

import (
	"context"
	"errors"
)

var (
	// ErrInvalidPlayer reports a missing player identifier.
	ErrInvalidPlayer = errors.New("invalid settings player")
	// ErrInvalidVolume reports a volume outside Nitro's supported range.
	ErrInvalidVolume = errors.New("invalid settings volume")
	// ErrInvalidHomeRoom reports a negative home-room identifier.
	ErrInvalidHomeRoom = errors.New("invalid home room")
	// ErrSettingsConflict reports an optimistic version mismatch.
	ErrSettingsConflict = errors.New("settings version conflict")
)

// Service validates and persists client settings outside hot paths.
type Service struct {
	// store persists settings records.
	store Store
}

// UpdateAdmin validates and applies one protected settings mutation.
func (service *Service) UpdateAdmin(ctx context.Context, playerID int64, expectedVersion int64, patch AdminPatch) (Record, error) {
	if playerID <= 0 || expectedVersion <= 0 {
		return Record{}, ErrInvalidPlayer
	}
	for _, value := range []*int32{patch.VolumeSystem, patch.VolumeFurniture, patch.VolumeTrax} {
		if value != nil && !validVolume(*value) {
			return Record{}, ErrInvalidVolume
		}
	}
	store, ok := service.store.(AdminStore)
	if !ok {
		return Record{}, ErrSettingsConflict
	}
	return store.UpdateAdmin(ctx, playerID, expectedVersion, patch)
}

// New creates a settings service.
func New(store Store) *Service { return &Service{store: store} }

// Find returns one player's settings.
func (service *Service) Find(ctx context.Context, playerID int64) (Record, error) {
	if playerID <= 0 {
		return Record{}, ErrInvalidPlayer
	}
	return service.store.Find(ctx, playerID)
}

// SetVolume validates and replaces all volume fields.
func (service *Service) SetVolume(ctx context.Context, playerID int64, system int32, furniture int32, trax int32) (Record, error) {
	if playerID <= 0 {
		return Record{}, ErrInvalidPlayer
	}
	if !validVolume(system) || !validVolume(furniture) || !validVolume(trax) {
		return Record{}, ErrInvalidVolume
	}
	return service.store.SetVolume(ctx, playerID, system, furniture, trax)
}

// SetOldChat replaces legacy chat selection.
func (service *Service) SetOldChat(ctx context.Context, playerID int64, oldChat bool) (Record, error) {
	if playerID <= 0 {
		return Record{}, ErrInvalidPlayer
	}
	return service.store.SetOldChat(ctx, playerID, oldChat)
}

// SetCameraFollowBlocked replaces camera-follow privacy.
func (service *Service) SetCameraFollowBlocked(ctx context.Context, playerID int64, blocked bool) (Record, error) {
	if playerID <= 0 {
		return Record{}, ErrInvalidPlayer
	}
	return service.store.SetCameraFollowBlocked(ctx, playerID, blocked)
}

// SetHomeRoom replaces or clears one player's home room.
func (service *Service) SetHomeRoom(ctx context.Context, playerID int64, roomID *int64) error {
	if playerID <= 0 {
		return ErrInvalidPlayer
	}
	if roomID != nil && *roomID <= 0 {
		return ErrInvalidHomeRoom
	}
	return service.store.SetHomeRoom(ctx, playerID, roomID)
}

// validVolume reports whether a value is supported by Nitro.
func validVolume(value int32) bool { return value >= 0 && value <= 100 }
