package service

import (
	"context"
	"errors"
	"fmt"

	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	"github.com/niflaot/pixels/internal/realm/player/repository"
)

// UpdateParams contains optional administrative player changes.
type UpdateParams struct {
	// Username replaces the visible name when present.
	Username *string
	// Look replaces the Nitro avatar figure when present.
	Look *string
	// Gender replaces the Nitro avatar gender when present.
	Gender *playermodel.Gender
	// Motto replaces the public motto when present.
	Motto *string
	// HomeRoomID replaces or clears the home room when present.
	HomeRoomID **int64
	// AllowNameChange replaces the username-change flag when present.
	AllowNameChange *bool
	// BubbleStyle replaces the selected Nitro bubble when present.
	BubbleStyle *int32
	// BlockFriendRequests replaces the friend-request privacy flag when present.
	BlockFriendRequests *bool
	// BlockRoomInvites replaces the room-invite privacy flag when present.
	BlockRoomInvites *bool
	// BlockFollowing replaces the follow privacy flag when present.
	BlockFollowing *bool
}

// Update applies one partial player identity and profile mutation atomically.
func (service *Service) Update(ctx context.Context, playerID int64, params UpdateParams) (Record, error) {
	if err := validatePlayerID(playerID); err != nil {
		return Record{}, err
	}
	if service.admin == nil {
		return Record{}, ErrAdminWriterUnavailable
	}
	record, found, err := service.FindByID(ctx, playerID)
	if err != nil {
		return Record{}, err
	}
	if !found {
		return Record{}, ErrPlayerNotFound
	}

	identityChanged, profileChanged, err := applyUpdate(&record, params)
	if err != nil {
		return Record{}, err
	}
	if !identityChanged && !profileChanged {
		return record, nil
	}

	err = service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		if identityChanged {
			updated, matched, updateErr := service.admin.UpdatePlayer(txCtx, repository.UpdatePlayerParams{
				PlayerID: playerID, Username: record.Player.Username, ExpectedVersion: record.Player.Version.Version,
			})
			if updateErr != nil {
				return updateErr
			}
			if !matched {
				return ErrConflict
			}
			record.Player = updated
		}
		if profileChanged {
			updated, matched, updateErr := service.admin.UpdateProfile(txCtx, updateProfileParams(record.Profile))
			if updateErr != nil {
				return updateErr
			}
			if !matched {
				return ErrConflict
			}
			record.Profile = updated
		}
		return nil
	})
	if errors.Is(err, repository.ErrUsernameTaken) {
		return Record{}, ErrUsernameTaken
	}
	if err != nil {
		return Record{}, fmt.Errorf("update player: %w", err)
	}

	return record, nil
}

// SoftDelete marks one player deleted so active lookups and future logins reject it.
func (service *Service) SoftDelete(ctx context.Context, playerID int64) error {
	if err := validatePlayerID(playerID); err != nil {
		return err
	}
	if service.admin == nil {
		return ErrAdminWriterUnavailable
	}
	record, found, err := service.FindByID(ctx, playerID)
	if err != nil {
		return err
	}
	if !found {
		return ErrPlayerNotFound
	}
	deleted, err := service.admin.SoftDeletePlayer(ctx, playerID, record.Player.Version.Version)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrConflict
	}

	return nil
}

// applyUpdate applies and validates optional fields in memory.
func applyUpdate(record *Record, params UpdateParams) (bool, bool, error) {
	identityChanged := params.Username != nil
	profileChanged := params.Look != nil || params.Gender != nil || params.Motto != nil || params.HomeRoomID != nil ||
		params.AllowNameChange != nil || params.BubbleStyle != nil || params.BlockFriendRequests != nil ||
		params.BlockRoomInvites != nil || params.BlockFollowing != nil
	if params.Username != nil {
		record.Player.Username = normalizeUsername(*params.Username)
		if err := validateUsername(record.Player.Username); err != nil {
			return false, false, err
		}
	}
	applyProfileUpdate(&record.Profile, params)
	if err := validateStoredProfile(record.Profile); err != nil {
		return false, false, err
	}

	return identityChanged, profileChanged, nil
}

// applyProfileUpdate replaces optional profile fields.
func applyProfileUpdate(profile *playermodel.Profile, params UpdateParams) {
	if params.Look != nil {
		profile.Look = *params.Look
	}
	if params.Gender != nil {
		profile.Gender = *params.Gender
	}
	if params.Motto != nil {
		profile.Motto = *params.Motto
	}
	if params.HomeRoomID != nil {
		profile.HomeRoomID = *params.HomeRoomID
	}
	if params.AllowNameChange != nil {
		profile.AllowNameChange = *params.AllowNameChange
	}
	if params.BubbleStyle != nil {
		profile.BubbleStyle = *params.BubbleStyle
	}
	if params.BlockFriendRequests != nil {
		profile.BlockFriendRequests = *params.BlockFriendRequests
	}
	if params.BlockRoomInvites != nil {
		profile.BlockRoomInvites = *params.BlockRoomInvites
	}
	if params.BlockFollowing != nil {
		profile.BlockFollowing = *params.BlockFollowing
	}
}

// validateStoredProfile validates one complete persisted profile candidate.
func validateStoredProfile(profile playermodel.Profile) error {
	if profile.BubbleStyle < 0 {
		return ErrInvalidBubbleStyle
	}
	if profile.HomeRoomID != nil && *profile.HomeRoomID <= 0 {
		return ErrInvalidHomeRoomID
	}
	return validateProfile(CreateProfileParams{Look: profile.Look, Gender: profile.Gender, Motto: profile.Motto, HomeRoomID: profile.HomeRoomID})
}

// updateProfileParams maps one profile into the repository replacement contract.
func updateProfileParams(profile playermodel.Profile) repository.UpdateProfileParams {
	return repository.UpdateProfileParams{CreateProfileParams: repository.CreateProfileParams{
		PlayerID: profile.PlayerID, Look: profile.Look, Gender: profile.Gender, Motto: profile.Motto,
		HomeRoomID: profile.HomeRoomID, AllowNameChange: profile.AllowNameChange,
	}, BubbleStyle: profile.BubbleStyle, BlockFriendRequests: profile.BlockFriendRequests,
		BlockRoomInvites: profile.BlockRoomInvites, BlockFollowing: profile.BlockFollowing,
		ExpectedVersion: profile.Version.Version}
}
