package session

import (
	"context"
	"testing"
	"time"

	navservice "github.com/niflaot/pixels/internal/realm/navigator/core"
	navrecord "github.com/niflaot/pixels/internal/realm/navigator/record"
)

// preferenceManager records final coalesced preferences.
type preferenceManager struct {
	navservice.Manager
	// saved stores persisted final values.
	saved []navrecord.Preference
}

// SavePreference records one persisted preference.
func (manager *preferenceManager) SavePreference(_ context.Context, preference navrecord.Preference) (navrecord.Preference, error) {
	manager.saved = append(manager.saved, preference)
	return preference, nil
}

// TestPreferenceWriterCoalescesAndBoundsPlayers verifies last-write-wins behavior.
func TestPreferenceWriterCoalescesAndBoundsPlayers(t *testing.T) {
	manager := &preferenceManager{}
	writer := NewPreferenceWriter(manager, nil, time.Second, 1)
	if !writer.Enqueue(navrecord.Preference{PlayerID: 1, WindowWidth: 320}) || !writer.Enqueue(navrecord.Preference{PlayerID: 1, WindowWidth: 640}) {
		t.Fatal("expected same player replacement")
	}
	if writer.Enqueue(navrecord.Preference{PlayerID: 2, WindowWidth: 320}) {
		t.Fatal("expected distinct player limit")
	}
	writer.flush(context.Background())
	if len(manager.saved) != 1 || manager.saved[0].WindowWidth != 640 {
		t.Fatalf("unexpected saved preferences %#v", manager.saved)
	}
}
