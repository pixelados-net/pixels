package service

import (
	"context"
	"errors"
	"testing"

	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	"github.com/niflaot/pixels/internal/realm/player/repository"
)

// adminStoreForTest adds administrative writes to the shared service fixture.
type adminStoreForTest struct {
	*fakeStore
	// conflict makes optimistic writes report no matching row.
	conflict bool
	// deleted reports that soft deletion was requested.
	deleted bool
	// playerUpdateErr fails identity writes.
	playerUpdateErr error
	// profileUpdateErr fails profile writes.
	profileUpdateErr error
	// deleteErr fails soft deletion.
	deleteErr error
}

// UpdatePlayer updates the fixture identity.
func (store *adminStoreForTest) UpdatePlayer(_ context.Context, params repository.UpdatePlayerParams) (playermodel.Player, bool, error) {
	if store.playerUpdateErr != nil {
		return playermodel.Player{}, false, store.playerUpdateErr
	}
	if store.conflict {
		return playermodel.Player{}, false, nil
	}
	store.player.Username = params.Username
	store.player.Version.Version++
	return store.player, true, nil
}

// UpdateProfile updates the fixture profile.
func (store *adminStoreForTest) UpdateProfile(_ context.Context, params repository.UpdateProfileParams) (playermodel.Profile, bool, error) {
	if store.profileUpdateErr != nil {
		return playermodel.Profile{}, false, store.profileUpdateErr
	}
	if store.conflict {
		return playermodel.Profile{}, false, nil
	}
	store.profile.Look = params.Look
	store.profile.Gender = params.Gender
	store.profile.Motto = params.Motto
	store.profile.HomeRoomID = params.HomeRoomID
	store.profile.AllowNameChange = params.AllowNameChange
	store.profile.BubbleStyle = params.BubbleStyle
	store.profile.BlockFriendRequests = params.BlockFriendRequests
	store.profile.BlockRoomInvites = params.BlockRoomInvites
	store.profile.BlockFollowing = params.BlockFollowing
	store.profile.Version.Version++
	return store.profile, true, nil
}

// SoftDeletePlayer records one fixture soft deletion.
func (store *adminStoreForTest) SoftDeletePlayer(context.Context, int64, int64) (bool, error) {
	if store.deleteErr != nil {
		return false, store.deleteErr
	}
	if store.conflict {
		return false, nil
	}
	store.deleted = true
	return true, nil
}

// TestUpdateAppliesIdentityAndProfileChanges verifies complete admin coordination.
func TestUpdateAppliesIdentityAndProfileChanges(t *testing.T) {
	store := &adminStoreForTest{fakeStore: newFakeStore()}
	username := "renamed"
	motto := "new motto"
	bubble := int32(4)
	record, err := New(store).Update(context.Background(), 7, UpdateParams{
		Username: &username, Motto: &motto, BubbleStyle: &bubble,
	})
	if err != nil {
		t.Fatalf("update player: %v", err)
	}
	if record.Player.Username != username || record.Profile.Motto != motto || record.Profile.BubbleStyle != bubble {
		t.Fatalf("unexpected record %#v", record)
	}
}

// TestUpdateRejectsOptimisticConflict verifies stale writes are explicit.
func TestUpdateRejectsOptimisticConflict(t *testing.T) {
	store := &adminStoreForTest{fakeStore: newFakeStore(), conflict: true}
	username := "renamed"
	_, err := New(store).Update(context.Background(), 7, UpdateParams{Username: &username})
	if !errors.Is(err, ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
}

// TestSoftDeleteMarksActivePlayer verifies service deletion behavior.
func TestSoftDeleteMarksActivePlayer(t *testing.T) {
	store := &adminStoreForTest{fakeStore: newFakeStore()}
	if err := New(store).SoftDelete(context.Background(), 7); err != nil {
		t.Fatalf("soft delete player: %v", err)
	}
	if !store.deleted {
		t.Fatal("expected soft deletion")
	}
}

// TestUpdateRejectsInvalidChanges verifies partial mutations use creation constraints.
func TestUpdateRejectsInvalidChanges(t *testing.T) {
	emptyUsername := " "
	negativeBubble := int32(-1)
	invalidHomeRoom := int64(0)
	homeRoom := &invalidHomeRoom
	tests := []struct {
		name     string
		params   UpdateParams
		expected error
	}{
		{name: "username", params: UpdateParams{Username: &emptyUsername}, expected: ErrInvalidUsername},
		{name: "bubble", params: UpdateParams{BubbleStyle: &negativeBubble}, expected: ErrInvalidBubbleStyle},
		{name: "home room", params: UpdateParams{HomeRoomID: &homeRoom}, expected: ErrInvalidHomeRoomID},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := New(&adminStoreForTest{fakeStore: newFakeStore()}).Update(context.Background(), 7, test.params)
			if !errors.Is(err, test.expected) {
				t.Fatalf("expected %v, got %v", test.expected, err)
			}
		})
	}
}

// TestUpdateHandlesAdministrativeStoreOutcomes verifies expected service mappings.
func TestUpdateHandlesAdministrativeStoreOutcomes(t *testing.T) {
	username := "renamed"
	t.Run("no changes", func(t *testing.T) {
		record, err := New(&adminStoreForTest{fakeStore: newFakeStore()}).Update(context.Background(), 7, UpdateParams{})
		if err != nil || record.Player.Username != "ian" {
			t.Fatalf("unexpected no-op record=%#v err=%v", record, err)
		}
	})
	t.Run("writer unavailable", func(t *testing.T) {
		_, err := New(newFakeStore()).Update(context.Background(), 7, UpdateParams{Username: &username})
		if !errors.Is(err, ErrAdminWriterUnavailable) {
			t.Fatalf("expected unavailable writer, got %v", err)
		}
	})
	t.Run("missing player", func(t *testing.T) {
		store := &adminStoreForTest{fakeStore: newFakeStore()}
		store.playerFound = false
		_, err := New(store).Update(context.Background(), 7, UpdateParams{Username: &username})
		if !errors.Is(err, ErrPlayerNotFound) {
			t.Fatalf("expected missing player, got %v", err)
		}
	})
	t.Run("duplicate username", func(t *testing.T) {
		store := &adminStoreForTest{fakeStore: newFakeStore(), playerUpdateErr: repository.ErrUsernameTaken}
		_, err := New(store).Update(context.Background(), 7, UpdateParams{Username: &username})
		if !errors.Is(err, ErrUsernameTaken) {
			t.Fatalf("expected duplicate username, got %v", err)
		}
	})
}

// TestSoftDeleteReportsExpectedFailures verifies deletion validation and conflicts.
func TestSoftDeleteReportsExpectedFailures(t *testing.T) {
	if err := New(newFakeStore()).SoftDelete(context.Background(), 0); !errors.Is(err, ErrInvalidPlayerID) {
		t.Fatalf("expected invalid id, got %v", err)
	}
	if err := New(newFakeStore()).SoftDelete(context.Background(), 7); !errors.Is(err, ErrAdminWriterUnavailable) {
		t.Fatalf("expected unavailable writer, got %v", err)
	}
	missing := &adminStoreForTest{fakeStore: newFakeStore()}
	missing.playerFound = false
	if err := New(missing).SoftDelete(context.Background(), 7); !errors.Is(err, ErrPlayerNotFound) {
		t.Fatalf("expected missing player, got %v", err)
	}
	conflict := &adminStoreForTest{fakeStore: newFakeStore(), conflict: true}
	if err := New(conflict).SoftDelete(context.Background(), 7); !errors.Is(err, ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
	expected := errors.New("database failed")
	failing := &adminStoreForTest{fakeStore: newFakeStore(), deleteErr: expected}
	if err := New(failing).SoftDelete(context.Background(), 7); !errors.Is(err, expected) {
		t.Fatalf("expected database error, got %v", err)
	}
}
