package bubble

import (
	"context"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/permission"
	permissionmodel "github.com/niflaot/pixels/internal/permission/model"
	bubblerepo "github.com/niflaot/pixels/internal/realm/chat/bubble/repository"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
)

// unlockStoreForTest stores bubble thresholds.
type unlockStoreForTest struct{ weights map[int32]int32 }

// List returns configured thresholds.
func (store *unlockStoreForTest) List(context.Context) ([]bubblerepo.Unlock, error) {
	items := make([]bubblerepo.Unlock, 0, len(store.weights))
	for id, weight := range store.weights {
		items = append(items, bubblerepo.Unlock{BubbleID: id, MinWeight: weight})
	}
	return items, nil
}

// MinWeight returns one threshold.
func (store *unlockStoreForTest) MinWeight(_ context.Context, id int32) (int32, bool, error) {
	weight, found := store.weights[id]
	return weight, found, nil
}

// Set stores one threshold.
func (store *unlockStoreForTest) Set(_ context.Context, id int32, weight int32) error {
	store.weights[id] = weight
	return nil
}

// Delete removes one threshold.
func (store *unlockStoreForTest) Delete(_ context.Context, id int32) error {
	delete(store.weights, id)
	return nil
}

// profileStoreForTest records bubble selections.
type profileStoreForTest struct{ style int32 }

// UpdateBubbleStyle records one selection.
func (store *profileStoreForTest) UpdateBubbleStyle(_ context.Context, playerID int64, style int32) (playermodel.Profile, error) {
	store.style = style
	return playermodel.Profile{PlayerID: playerID, BubbleStyle: style}, nil
}

// permissionsForTest resolves one weight and bypass.
type permissionsForTest struct {
	// bypass reports unrestricted bubble access.
	bypass bool
	// weight stores the primary group weight.
	weight int32
}

// HasPermission reports unrestricted access.
func (value permissionsForTest) HasPermission(context.Context, int64, permission.Node) (bool, error) {
	return value.bypass, nil
}

// PrimaryGroup returns one weighted group.
func (value permissionsForTest) PrimaryGroup(context.Context, int64) (permissionmodel.Group, bool, error) {
	return permissionmodel.Group{Weight: value.weight}, true, nil
}

// TestAllowedEnforcesWeightAndBypass verifies bubble unlock policy.
func TestAllowedEnforcesWeightAndBypass(t *testing.T) {
	store := &unlockStoreForTest{weights: map[int32]int32{5: 50}}
	locked := New(store, &profileStoreForTest{}, permissionsForTest{weight: 20}, playerlive.NewRegistry(), "bubble.any")
	if allowed, err := locked.Allowed(context.Background(), 1, 5); err != nil || allowed {
		t.Fatalf("allowed=%v err=%v", allowed, err)
	}
	unlocked := New(store, &profileStoreForTest{}, permissionsForTest{bypass: true}, playerlive.NewRegistry(), "bubble.any")
	if allowed, err := unlocked.Allowed(context.Background(), 1, 5); err != nil || !allowed {
		t.Fatalf("allowed=%v err=%v", allowed, err)
	}
}

// TestSelectPersistsAllowedStyle verifies validated profile persistence.
func TestSelectPersistsAllowedStyle(t *testing.T) {
	profiles := &profileStoreForTest{}
	service := New(&unlockStoreForTest{weights: map[int32]int32{}}, profiles, permissionsForTest{}, playerlive.NewRegistry(), "bubble.any")
	if err := service.Select(context.Background(), 1, 4); err != nil {
		t.Fatalf("select: %v", err)
	}
	if profiles.style != 4 {
		t.Fatalf("expected style 4, got %d", profiles.style)
	}
	if err := service.SetUnlock(context.Background(), -1, 0); err == nil {
		t.Fatal("expected invalid threshold")
	}
}

// TestSelectUpdatesLiveSnapshotAndAdminMutations verifies runtime projection and thresholds.
func TestSelectUpdatesLiveSnapshotAndAdminMutations(t *testing.T) {
	unlocks := &unlockStoreForTest{weights: map[int32]int32{7: 10}}
	players := playerlive.NewRegistry()
	peer, _ := playerlive.NewSessionPeer("connection", "websocket", time.Now())
	player, _ := playerlive.NewPlayer(playerlive.Snapshot{ID: 1, Username: "alice"}, peer)
	_ = players.Add(player)
	service := New(unlocks, &profileStoreForTest{}, permissionsForTest{weight: 10}, players, "bubble.any")
	if err := service.Select(context.Background(), 1, 7); err != nil {
		t.Fatalf("select: %v", err)
	}
	if player.Snapshot().BubbleStyle != 7 {
		t.Fatalf("expected live style 7, got %d", player.Snapshot().BubbleStyle)
	}
	if err := service.SetUnlock(context.Background(), 8, 20); err != nil {
		t.Fatalf("set unlock: %v", err)
	}
	items, err := service.List(context.Background())
	if err != nil || len(items) != 2 {
		t.Fatalf("items=%#v err=%v", items, err)
	}
	if err = service.DeleteUnlock(context.Background(), 8); err != nil {
		t.Fatalf("delete unlock: %v", err)
	}
}

// TestSelectRejectsLockedAndInvalidStyles verifies expected selection failures.
func TestSelectRejectsLockedAndInvalidStyles(t *testing.T) {
	service := New(&unlockStoreForTest{weights: map[int32]int32{5: 50}}, &profileStoreForTest{}, permissionsForTest{weight: 10}, playerlive.NewRegistry(), "bubble.any")
	if err := service.Select(context.Background(), 1, 5); err != ErrBubbleLocked {
		t.Fatalf("expected locked error, got %v", err)
	}
	if _, err := service.Allowed(context.Background(), 1, -1); err != ErrInvalidBubble {
		t.Fatalf("expected invalid error, got %v", err)
	}
	if err := service.DeleteUnlock(context.Background(), -1); err != ErrInvalidBubble {
		t.Fatalf("expected invalid delete, got %v", err)
	}
}
