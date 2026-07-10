package service

import (
	"context"
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

	player, err := service.store.CreatePlayer(ctx, repository.CreatePlayerParams{Username: username})
	if err != nil {
		return Record{}, fmt.Errorf("create player identity: %w", err)
	}

	profile, err := service.store.CreateProfile(ctx, profileParams(player.ID, params.Profile))
	if err != nil {
		return Record{}, fmt.Errorf("create player profile: %w", err)
	}
	if service.permissions != nil {
		if err := service.permissions.AssignDefaultGroup(ctx, player.ID); err != nil {
			return Record{}, fmt.Errorf("assign player default permission group: %w", err)
		}
	}

	return Record{Player: player, Profile: profile}, nil
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
