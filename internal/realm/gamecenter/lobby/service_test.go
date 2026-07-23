package lobby

import (
	"context"
	"testing"

	gamecenterconfig "github.com/niflaot/pixels/internal/realm/gamecenter/config"
	gamecenterrecord "github.com/niflaot/pixels/internal/realm/gamecenter/record"
)

// memoryStore provides deterministic service tests.
type memoryStore struct{ games []gamecenterrecord.Game }

// ListGames returns configured test games.
func (store memoryStore) ListGames(context.Context, bool) ([]gamecenterrecord.Game, error) {
	return append([]gamecenterrecord.Game(nil), store.games...), nil
}

// FindGame returns no direct test lookup.
func (memoryStore) FindGame(context.Context, int32) (gamecenterrecord.Game, bool, error) {
	return gamecenterrecord.Game{}, false, nil
}

// UpsertScore accepts test scores.
func (memoryStore) UpsertScore(context.Context, int32, int64, int32, int32, int64) error { return nil }

// TestServiceReloadAndLaunch verifies cache replacement and honest missing-URL behavior.
func TestServiceReloadAndLaunch(t *testing.T) {
	service := New(gamecenterconfig.Config{Enabled: true}, memoryStore{games: []gamecenterrecord.Game{{ID: 2, Name: "test", Enabled: true, LaunchURL: "https://game"}}})
	if err := service.Reload(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(service.List()) != 1 {
		t.Fatal("expected one game")
	}
	if _, err := service.FindLaunch(2); err != nil {
		t.Fatal(err)
	}
	if _, err := service.FindLaunch(3); err == nil {
		t.Fatal("expected unavailable game")
	}
}
