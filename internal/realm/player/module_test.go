package player

import (
	"testing"

	playerauthenticated "github.com/niflaot/pixels/internal/realm/player/events/authenticated"
	playerauthenticating "github.com/niflaot/pixels/internal/realm/player/events/authenticating"
	playerauthfailed "github.com/niflaot/pixels/internal/realm/player/events/authfailed"
	playerconnected "github.com/niflaot/pixels/internal/realm/player/events/connected"
	playerdisconnected "github.com/niflaot/pixels/internal/realm/player/events/disconnected"
	playerprofileloaded "github.com/niflaot/pixels/internal/realm/player/events/profileloaded"
	"github.com/niflaot/pixels/internal/realm/player/service"
)

// TestEventNames verifies player event names are stable.
func TestEventNames(t *testing.T) {
	events := []string{
		string(playerauthenticating.Name),
		string(playerauthenticated.Name),
		string(playerauthfailed.Name),
		string(playerconnected.Name),
		string(playerdisconnected.Name),
		string(playerprofileloaded.Name),
	}

	for _, event := range events {
		if event == "" {
			t.Fatal("expected event name")
		}
	}
}

// TestProvidersExposeContracts verifies module helper providers return contracts.
func TestProvidersExposeContracts(t *testing.T) {
	playerService := service.New(nil)

	if NewStore(nil) == nil {
		t.Fatal("expected store")
	}
	if NewCreator(playerService) == nil {
		t.Fatal("expected creator")
	}
	if NewFinder(playerService) == nil {
		t.Fatal("expected finder")
	}
	if NewManager(playerService) == nil {
		t.Fatal("expected manager")
	}
}
