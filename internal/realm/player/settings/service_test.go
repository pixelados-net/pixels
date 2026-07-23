package settings

import (
	"context"
	"errors"
	"testing"
)

// memoryStore stores settings for focused service tests.
type memoryStore struct{ record Record }

// Find returns the in-memory record.
func (store *memoryStore) Find(context.Context, int64) (Record, error) { return store.record, nil }

// SetVolume records volume fields.
func (store *memoryStore) SetVolume(_ context.Context, playerID int64, system int32, furniture int32, trax int32) (Record, error) {
	store.record = Record{PlayerID: playerID, VolumeSystem: system, VolumeFurniture: furniture, VolumeTrax: trax}
	return store.record, nil
}

// SetOldChat records old-chat selection.
func (store *memoryStore) SetOldChat(_ context.Context, playerID int64, oldChat bool) (Record, error) {
	store.record.PlayerID, store.record.OldChat = playerID, oldChat
	return store.record, nil
}

// SetCameraFollowBlocked records camera-follow privacy.
func (store *memoryStore) SetCameraFollowBlocked(_ context.Context, playerID int64, blocked bool) (Record, error) {
	store.record.PlayerID = playerID
	store.record.CameraFollowBlocked = blocked
	return store.record, nil
}

// SetHomeRoom accepts a home-room update.
func (*memoryStore) SetHomeRoom(context.Context, int64, *int64) error { return nil }

// UpdateAdmin applies one deterministic settings patch.
func (store *memoryStore) UpdateAdmin(_ context.Context, playerID int64, _ int64, patch AdminPatch) (Record, error) {
	store.record.PlayerID = playerID
	if patch.VolumeSystem != nil {
		store.record.VolumeSystem = *patch.VolumeSystem
	}
	store.record.Version++
	return store.record, nil
}

// TestSetVolumeValidatesBounds verifies server-authoritative volume ranges.
func TestSetVolumeValidatesBounds(t *testing.T) {
	service := New(&memoryStore{})
	if _, err := service.SetVolume(context.Background(), 1, 0, 50, 100); err != nil {
		t.Fatalf("set valid volume: %v", err)
	}
	if _, err := service.SetVolume(context.Background(), 1, -1, 0, 0); !errors.Is(err, ErrInvalidVolume) {
		t.Fatalf("expected invalid volume, got %v", err)
	}
}

// TestSettingsServiceValidatesAllOperations verifies player, home-room, and admin bounds.
func TestSettingsServiceValidatesAllOperations(t *testing.T) {
	store := &memoryStore{record: Default(1)}
	service := New(store)
	if _, err := service.Find(context.Background(), 0); !errors.Is(err, ErrInvalidPlayer) {
		t.Fatalf("expected invalid find, got %v", err)
	}
	if record, err := service.Find(context.Background(), 1); err != nil || record.VolumeSystem != 100 {
		t.Fatalf("record=%#v err=%v", record, err)
	}
	if _, err := service.SetOldChat(context.Background(), 0, true); !errors.Is(err, ErrInvalidPlayer) {
		t.Fatalf("expected invalid old chat, got %v", err)
	}
	roomID := int64(-1)
	if err := service.SetHomeRoom(context.Background(), 1, &roomID); !errors.Is(err, ErrInvalidHomeRoom) {
		t.Fatalf("expected invalid home room, got %v", err)
	}
	volume := int32(75)
	record, err := service.UpdateAdmin(context.Background(), 1, 1, AdminPatch{VolumeSystem: &volume})
	if err != nil || record.VolumeSystem != 75 {
		t.Fatalf("record=%#v err=%v", record, err)
	}
	invalid := int32(101)
	if _, err = service.UpdateAdmin(context.Background(), 1, 1, AdminPatch{VolumeSystem: &invalid}); !errors.Is(err, ErrInvalidVolume) {
		t.Fatalf("expected invalid admin volume, got %v", err)
	}
}

// TestSettingsConfigAndDefaults verifies environment and default projections.
func TestSettingsConfigAndDefaults(t *testing.T) {
	if record := Default(9); record.PlayerID != 9 || record.Version != 1 {
		t.Fatalf("default=%#v", record)
	}
	t.Setenv("PIXELS_PLAYER_SETTINGS_PENDING_LIMIT", "12")
	config, err := LoadConfig()
	if err != nil || config.PendingLimit != 12 {
		t.Fatalf("config=%#v err=%v", config, err)
	}
}
