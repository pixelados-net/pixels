package service

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	"github.com/niflaot/pixels/internal/realm/player/repository"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestCreateCreatesPlayerAndProfile verifies validated player creation.
func TestCreateCreatesPlayerAndProfile(t *testing.T) {
	store := newFakeStore()
	record, err := New(store).Create(context.Background(), CreateParams{
		Username: "  ian  ",
		Profile: CreateProfileParams{
			Look:  "hd-180-1",
			Motto: "hello",
		},
	})
	if err != nil {
		t.Fatalf("create player: %v", err)
	}

	if record.Player.Username != "ian" {
		t.Fatalf("expected normalized username, got %q", record.Player.Username)
	}

	if record.Profile.Gender != playermodel.GenderMale {
		t.Fatalf("expected default male gender, got %q", record.Profile.Gender)
	}
}

// TestCreateRejectsInvalidInput verifies creation validation.
func TestCreateRejectsInvalidInput(t *testing.T) {
	cases := []struct {
		name     string
		params   CreateParams
		expected error
	}{
		{name: "username", params: CreateParams{}, expected: ErrInvalidUsername},
		{name: "gender", params: CreateParams{Username: "ian", Profile: CreateProfileParams{Gender: playermodel.Gender("X")}}, expected: ErrInvalidGender},
		{name: "look", params: CreateParams{Username: "ian", Profile: CreateProfileParams{Look: strings.Repeat("x", MaxLookLength+1)}}, expected: ErrInvalidLook},
		{name: "motto", params: CreateParams{Username: "ian", Profile: CreateProfileParams{Motto: strings.Repeat("x", MaxMottoLength+1)}}, expected: ErrInvalidMotto},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, err := New(newFakeStore()).Create(context.Background(), test.params)
			if !errors.Is(err, test.expected) {
				t.Fatalf("expected %v, got %v", test.expected, err)
			}
		})
	}
}

// TestFindByIDLoadsPlayerAndProfile verifies lookup composition.
func TestFindByIDLoadsPlayerAndProfile(t *testing.T) {
	store := newFakeStore()
	record, found, err := New(store).FindByID(context.Background(), 7)
	if err != nil {
		t.Fatalf("find player: %v", err)
	}

	if !found {
		t.Fatal("expected player")
	}

	if record.Profile.PlayerID != 7 {
		t.Fatalf("expected profile player id 7, got %d", record.Profile.PlayerID)
	}
}

// TestFindByUsernameNormalizesInput verifies username lookup normalization.
func TestFindByUsernameNormalizesInput(t *testing.T) {
	store := newFakeStore()
	_, found, err := New(store).FindByUsername(context.Background(), " IAN ")
	if err != nil {
		t.Fatalf("find player: %v", err)
	}

	if !found {
		t.Fatal("expected player")
	}

	if store.lastUsername != "IAN" {
		t.Fatalf("expected normalized username, got %q", store.lastUsername)
	}
}

// TestFindByIDRejectsInvalidID verifies id validation.
func TestFindByIDRejectsInvalidID(t *testing.T) {
	_, _, err := New(newFakeStore()).FindByID(context.Background(), 0)
	if !errors.Is(err, ErrInvalidPlayerID) {
		t.Fatalf("expected invalid player id, got %v", err)
	}
}

// TestFindByUsernameReturnsMissingPlayer verifies missing player lookup.
func TestFindByUsernameReturnsMissingPlayer(t *testing.T) {
	store := newFakeStore()
	store.playerFound = false

	_, found, err := New(store).FindByUsername(context.Background(), "ian")
	if err != nil {
		t.Fatalf("find player: %v", err)
	}

	if found {
		t.Fatal("expected missing player")
	}
}

// TestFindByIDReturnsMissingProfile verifies incomplete records are missing.
func TestFindByIDReturnsMissingProfile(t *testing.T) {
	store := newFakeStore()
	store.profileFound = false

	_, found, err := New(store).FindByID(context.Background(), 7)
	if err != nil {
		t.Fatalf("find player: %v", err)
	}

	if found {
		t.Fatal("expected missing composed record")
	}
}

// TestFindByIDWrapsProfileError verifies profile lookup errors are wrapped.
func TestFindByIDWrapsProfileError(t *testing.T) {
	expected := errors.New("database failed")
	store := newFakeStore()
	store.profileErr = expected

	_, _, err := New(store).FindByID(context.Background(), 7)
	if !errors.Is(err, expected) {
		t.Fatalf("expected profile error, got %v", err)
	}
}

// newFakeStore creates a store with default records.
func newFakeStore() *fakeStore {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)

	return &fakeStore{
		playerFound:  true,
		profileFound: true,
		player: playermodel.Player{
			Base: sharedmodel.Base{
				Identity:   sharedmodel.Identity{ID: 7},
				Timestamps: sharedmodel.Timestamps{CreatedAt: now, UpdatedAt: now},
				Version:    sharedmodel.Version{Version: 1},
			},
			Username: "ian",
		},
		profile: playermodel.Profile{
			PlayerID: 7,
			Look:     "hd-180-1",
			Gender:   playermodel.GenderMale,
			Motto:    "hello",
			Timestamps: sharedmodel.Timestamps{
				CreatedAt: now,
				UpdatedAt: now,
			},
			Version: sharedmodel.Version{Version: 1},
		},
	}
}

// fakeStore records player store calls for tests.
type fakeStore struct {
	// player is the returned player.
	player playermodel.Player

	// profile is the returned profile.
	profile playermodel.Profile

	// playerFound reports whether player lookups succeed.
	playerFound bool

	// profileFound reports whether profile lookups succeed.
	profileFound bool

	// playerErr is returned by player lookups.
	playerErr error

	// profileErr is returned by profile lookups.
	profileErr error

	// lastUsername is the last username lookup value.
	lastUsername string
}

// WithinTransaction runs test work synchronously.
func (store *fakeStore) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	return work(ctx)
}

// CreatePlayer creates a player identity record for tests.
func (store *fakeStore) CreatePlayer(_ context.Context, params repository.CreatePlayerParams) (playermodel.Player, error) {
	store.player.Username = params.Username

	return store.player, store.playerErr
}

// FindPlayerByID finds an active player by id for tests.
func (store *fakeStore) FindPlayerByID(context.Context, int64) (playermodel.Player, bool, error) {
	return store.player, store.playerFound, store.playerErr
}

// FindPlayerByUsername finds an active player by username for tests.
func (store *fakeStore) FindPlayerByUsername(_ context.Context, username string) (playermodel.Player, bool, error) {
	store.lastUsername = username

	return store.player, store.playerFound, store.playerErr
}

// CreateProfile creates a player profile record for tests.
func (store *fakeStore) CreateProfile(_ context.Context, params repository.CreateProfileParams) (playermodel.Profile, error) {
	store.profile.PlayerID = params.PlayerID
	store.profile.Look = params.Look
	store.profile.Gender = params.Gender
	store.profile.Motto = params.Motto
	store.profile.HomeRoomID = params.HomeRoomID
	store.profile.AllowNameChange = params.AllowNameChange

	return store.profile, store.profileErr
}

// UpdateBubbleStyle persists one bubble style for tests.
func (store *fakeStore) UpdateBubbleStyle(_ context.Context, _ int64, bubbleStyle int32) (playermodel.Profile, error) {
	store.profile.BubbleStyle = bubbleStyle

	return store.profile, store.profileErr
}

// UpdatePrivacy persists messenger privacy for tests.
func (store *fakeStore) UpdatePrivacy(_ context.Context, _ int64, params repository.PrivacyParams) (playermodel.Profile, error) {
	store.profile.BlockFriendRequests = params.BlockFriendRequests
	store.profile.BlockRoomInvites = params.BlockRoomInvites
	store.profile.BlockFollowing = params.BlockFollowing

	return store.profile, store.profileErr
}

// FindProfileByPlayerID finds a profile by player id for tests.
func (store *fakeStore) FindProfileByPlayerID(context.Context, int64) (playermodel.Profile, bool, error) {
	return store.profile, store.profileFound, store.profileErr
}
