package identity

import (
	"context"
	"errors"
	"testing"

	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
)

// identityFinder stores exact usernames and rename policy for focused tests.
type identityFinder struct {
	// available reports whether the actor can rename.
	available bool
	// taken stores case-insensitive existing names.
	taken map[string]bool
}

// FindByID returns the actor policy.
func (finder identityFinder) FindByID(context.Context, int64) (playerservice.Record, bool, error) {
	return playerservice.Record{Profile: playermodel.Profile{AllowNameChange: finder.available}}, true, nil
}

// FindByUsername reports configured collisions.
func (finder identityFinder) FindByUsername(_ context.Context, username string) (playerservice.Record, bool, error) {
	return playerservice.Record{}, finder.taken[username], nil
}

// renameStore records committed candidates.
type renameStore struct{ candidate string }

// identityFilter reports a deterministic protected-word match.
type identityFilter struct{}

// Censor rejects candidates containing the test protected word.
func (identityFilter) Censor(value string) (string, bool) { return value, value == "blocked" }

// Rename stores one candidate.
func (store *renameStore) Rename(_ context.Context, _ int64, candidate string) (RenameResult, error) {
	store.candidate = candidate
	return RenameResult{OldUsername: "old", NewUsername: candidate}, nil
}

// TestCheckRejectsReservedAndFilteredNames verifies server-owned name policy.
func TestCheckRejectsReservedAndFilteredNames(t *testing.T) {
	service := NewConfigured(&renameStore{}, identityFinder{available: true, taken: map[string]bool{}}, nil, identityFilter{}, DefaultConfig())
	for _, candidate := range []string{"admin", "blocked"} {
		result, err := service.Check(context.Background(), 1, candidate)
		if err != nil || result.Code != ResultInvalid {
			t.Fatalf("candidate=%q result=%#v err=%v", candidate, result, err)
		}
	}
}

// TestValidateUsernameCodes verifies Nitro's complete validation matrix.
func TestValidateUsernameCodes(t *testing.T) {
	service := New(&renameStore{}, identityFinder{}, nil)
	cases := map[string]int32{"ab": ResultTooShort, "abcdefghijklmnop": ResultTooLong, "bad name": ResultInvalid, "Valid_1": ResultAvailable}
	for value, expected := range cases {
		if actual := service.validate(value); actual != expected {
			t.Fatalf("validate %q=%d expected %d", value, actual, expected)
		}
	}
}

// TestCheckHonorsPolicyAndProducesDeterministicSuggestions verifies availability behavior.
func TestCheckHonorsPolicyAndProducesDeterministicSuggestions(t *testing.T) {
	service := New(&renameStore{}, identityFinder{available: false, taken: map[string]bool{}}, nil)
	result, err := service.Check(context.Background(), 1, "Valid")
	if err != nil || result.Code != ResultDisabled {
		t.Fatalf("result=%#v err=%v", result, err)
	}
	service = New(&renameStore{}, identityFinder{available: true, taken: map[string]bool{"Valid": true}}, nil)
	result, err = service.Check(context.Background(), 1, "Valid")
	if err != nil || result.Code != ResultTaken || len(result.Suggestions) != 4 || result.Suggestions[0] != "Valid1" {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

// TestRenameValidatesAndCommits verifies the cold-path rename boundary.
func TestRenameValidatesAndCommits(t *testing.T) {
	store := &renameStore{}
	service := New(store, identityFinder{available: true, taken: map[string]bool{}}, nil)
	if _, err := service.Rename(context.Background(), 1, "bad name"); !errors.Is(err, ErrReservationMissing) {
		t.Fatalf("expected invalid rename, got %v", err)
	}
	if _, err := service.Rename(context.Background(), 1, "admin"); !errors.Is(err, ErrReservationMissing) {
		t.Fatalf("expected reserved rename rejection, got %v", err)
	}
	result, err := service.Rename(context.Background(), 1, "Valid")
	if err != nil || result.NewUsername != "Valid" || store.candidate != "Valid" {
		t.Fatalf("result=%#v candidate=%q err=%v", result, store.candidate, err)
	}
	if reservationKey("VaLiD") != "player:name:reservation:valid" {
		t.Fatal("unexpected reservation key")
	}
}

// TestConfiguredPolicyFallsBackAndLoadsEnvironment verifies config boundaries.
func TestConfiguredPolicyFallsBackAndLoadsEnvironment(t *testing.T) {
	service := NewConfigured(&renameStore{}, identityFinder{}, nil, nil, Config{})
	if service.config.MinimumLength != DefaultConfig().MinimumLength {
		t.Fatalf("config=%#v", service.config)
	}
	t.Setenv("PIXELS_PLAYER_USERNAME_MIN_LENGTH", "4")
	config, err := LoadConfig()
	if err != nil || config.MinimumLength != 4 {
		t.Fatalf("config=%#v err=%v", config, err)
	}
}
