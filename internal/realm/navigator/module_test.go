package navigator

import (
	"testing"

	navsearch "github.com/niflaot/pixels/internal/realm/navigator/browse/search/events/executed"
	"github.com/niflaot/pixels/internal/realm/navigator/core"
	navfavorite "github.com/niflaot/pixels/internal/realm/navigator/favorite/events/changed"
	navclosed "github.com/niflaot/pixels/internal/realm/navigator/session/events/closed"
	navinitialized "github.com/niflaot/pixels/internal/realm/navigator/session/events/initialized"
)

// TestEventNames verifies navigator event names are stable.
func TestEventNames(t *testing.T) {
	events := []string{
		string(navinitialized.Name),
		string(navclosed.Name),
		string(navsearch.Name),
		string(navfavorite.Name),
	}

	for _, event := range events {
		if event == "" {
			t.Fatal("expected event name")
		}
	}
}

// TestProvidersExposeContracts verifies module helper providers return contracts.
func TestProvidersExposeContracts(t *testing.T) {
	navigatorService := core.New(nil)

	if NewStore(nil) == nil {
		t.Fatal("expected store")
	}
	if NewManager(navigatorService) == nil {
		t.Fatal("expected navigator manager")
	}
}
