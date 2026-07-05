package player

import (
	"testing"

	"github.com/niflaot/pixels/internal/realm/player/service"
)

// TestEventNames verifies player event names are stable.
func TestEventNames(t *testing.T) {
	events := []string{
		string(EventAuthenticating),
		string(EventAuthenticated),
		string(EventAuthenticationFailed),
		string(EventConnected),
		string(EventDisconnected),
		string(EventProfileLoaded),
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
