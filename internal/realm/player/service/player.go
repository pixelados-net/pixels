package service

import (
	"context"
	"errors"
	"fmt"

	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	"github.com/niflaot/pixels/internal/realm/player/repository"
)

// CreateParams contains player creation input.
type CreateParams struct {
	// Username is the unique visible player name.
	Username string

	// Profile contains the player presentation input.
	Profile CreateProfileParams
}

// CreateProfileParams contains player profile creation input.
type CreateProfileParams struct {
	// Look is the Nitro avatar figure string.
	Look string

	// Gender is the Nitro avatar gender code.
	Gender playermodel.Gender

	// Motto is the public player motto.
	Motto string

	// HomeRoomID is the optional default home room identifier.
	HomeRoomID *int64

	// AllowNameChange reports whether the player can change username.
	AllowNameChange bool
}

// Create creates a player with a profile.
func (service *Service) Create(ctx context.Context, params CreateParams) (Record, error) {
	username := normalizeUsername(params.Username)
	if err := validateUsername(username); err != nil {
		return Record{}, err
	}

	params.Profile.Gender = defaultGender(params.Profile.Gender)
	if err := validateProfile(params.Profile); err != nil {
		return Record{}, err
	}

	var record Record
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		player, createErr := service.store.CreatePlayer(txCtx, repository.CreatePlayerParams{Username: username})
		if createErr != nil {
			return fmt.Errorf("create player identity: %w", createErr)
		}

		profile, createErr := service.store.CreateProfile(txCtx, profileParams(player.ID, params.Profile))
		if createErr != nil {
			return fmt.Errorf("create player profile: %w", createErr)
		}
		if service.permissions != nil {
			if createErr := service.permissions.AssignDefaultGroup(txCtx, player.ID); createErr != nil {
				return fmt.Errorf("assign player default permission group: %w", createErr)
			}
		}

		record = Record{Player: player, Profile: profile}
		return nil
	})
	if err != nil {
		if errors.Is(err, repository.ErrUsernameTaken) {
			return Record{}, ErrUsernameTaken
		}
		return Record{}, err
	}

	return record, nil
}

// FindByID finds a player by id.
func (service *Service) FindByID(ctx context.Context, id int64) (Record, bool, error) {
	if err := validatePlayerID(id); err != nil {
		return Record{}, false, err
	}

	player, found, err := service.store.FindPlayerByID(ctx, id)
	if err != nil || !found {
		return Record{Player: player}, found, err
	}

	return service.record(ctx, player)
}

// FindByUsername finds a player by username.
func (service *Service) FindByUsername(ctx context.Context, username string) (Record, bool, error) {
	username = normalizeUsername(username)
	if err := validateUsername(username); err != nil {
		return Record{}, false, err
	}

	player, found, err := service.store.FindPlayerByUsername(ctx, username)
	if err != nil || !found {
		return Record{Player: player}, found, err
	}

	return service.record(ctx, player)
}

// SetClub updates one player's derived club entitlement.
func (service *Service) SetClub(ctx context.Context, playerID int64, club playermodel.Club) error {
	if err := validatePlayerID(playerID); err != nil {
		return err
	}
	if service.clubs == nil {
		return ErrClubWriterUnavailable
	}

	return service.clubs.UpdateClub(ctx, playerID, club)
}

// UpdatePrivacy persists messenger privacy fields.
func (service *Service) UpdatePrivacy(ctx context.Context, playerID int64, params PrivacyParams) (Record, error) {
	player, found, err := service.store.FindPlayerByID(ctx, playerID)
	if err != nil {
		return Record{}, err
	}
	if !found {
		return Record{}, ErrPlayerNotFound
	}
	profile, err := service.store.UpdatePrivacy(ctx, playerID, repository.PrivacyParams{
		BlockFriendRequests: params.BlockFriendRequests,
		BlockRoomInvites:    params.BlockRoomInvites,
		BlockFollowing:      params.BlockFollowing,
	})
	if err != nil {
		return Record{}, fmt.Errorf("update player messenger privacy: %w", err)
	}

	return Record{Player: player, Profile: profile}, nil
}

// record loads a complete player record.
func (service *Service) record(ctx context.Context, player playermodel.Player) (Record, bool, error) {
	profile, found, err := service.store.FindProfileByPlayerID(ctx, player.ID)
	if err != nil {
		return Record{}, false, fmt.Errorf("find player profile: %w", err)
	}

	if !found {
		return Record{Player: player}, false, nil
	}

	return Record{Player: player, Profile: profile}, true, nil
}

// profileParams maps service profile input to repository input.
func profileParams(playerID int64, params CreateProfileParams) repository.CreateProfileParams {
	return repository.CreateProfileParams{
		PlayerID:        playerID,
		Look:            params.Look,
		Gender:          params.Gender,
		Motto:           params.Motto,
		HomeRoomID:      params.HomeRoomID,
		AllowNameChange: params.AllowNameChange,
	}
}
