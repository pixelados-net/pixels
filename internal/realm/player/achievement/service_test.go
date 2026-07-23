package achievement

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// achievementStore records badge and respect operations.
type achievementStore struct {
	// badges stores the durable fixture.
	badges []Badge
	// code stores the last badge code.
	code string
	// equipped stores the last active badge replacement.
	equipped []string
	// respectKey stores the last respect idempotency key.
	respectKey string
}

// Badges returns the durable fixture.
func (store *achievementStore) Badges(context.Context, int64) ([]Badge, error) {
	return store.badges, nil
}

// GrantBadge records one badge grant.
func (store *achievementStore) GrantBadge(_ context.Context, _ int64, code string, _ string) (bool, error) {
	store.code = code
	return true, nil
}

// ReplaceBadge records one badge code replacement.
func (store *achievementStore) ReplaceBadge(_ context.Context, _ int64, oldCode string, newCode string, _ string) (bool, error) {
	for index := range store.badges {
		if strings.EqualFold(store.badges[index].Code, oldCode) {
			store.badges[index].Code = newCode
			return true, nil
		}
	}
	return false, nil
}

// RemoveBadge records one badge removal.
func (store *achievementStore) RemoveBadge(_ context.Context, _ int64, code string) (bool, error) {
	for index := range store.badges {
		if strings.EqualFold(store.badges[index].Code, code) {
			store.badges = append(store.badges[:index], store.badges[index+1:]...)
			return true, nil
		}
	}
	return false, nil
}

// SetEquipped records one active badge replacement.
func (store *achievementStore) SetEquipped(_ context.Context, _ int64, codes []string) error {
	store.equipped = append([]string(nil), codes...)
	for index := range store.badges {
		store.badges[index].Equipped = false
		store.badges[index].Slot = 0
		for slot, code := range codes {
			if strings.EqualFold(store.badges[index].Code, code) {
				store.badges[index].Equipped = true
				store.badges[index].Slot = int32(slot + 1)
			}
		}
	}
	return nil
}

// GrantRespect records one respect grant.
func (store *achievementStore) GrantRespect(_ context.Context, _ int64, _ int32, key string, _ string) (bool, error) {
	store.respectKey = key
	return true, nil
}

// TestServiceProjectsEquippedBadgesAndValidatesGrants verifies hot snapshots and durable commands.
func TestServiceProjectsEquippedBadgesAndValidatesGrants(t *testing.T) {
	store := &achievementStore{badges: []Badge{{ID: 1, Code: "ADM", Equipped: true, Slot: 1}, {ID: 2, Code: "OWNED"}}}
	service := New(store)
	if err := service.Load(context.Background(), 8); err != nil {
		t.Fatal(err)
	}
	if wearing, loaded := service.Wearing(8, "ADM"); !wearing || !loaded {
		t.Fatalf("wearing=%v loaded=%v", wearing, loaded)
	}
	if wearing, _ := service.Wearing(8, "OWNED"); wearing {
		t.Fatal("owned but unequipped badge passed")
	}
	if _, err := service.GrantBadge(context.Background(), 8, " qa_badge ", "wired"); err != nil || store.code != "QA_BADGE" {
		t.Fatalf("badge code=%q err=%v", store.code, err)
	}
	if granted, err := service.GrantRespect(context.Background(), 8, 1001, "bad"); err != nil || granted || store.respectKey != "" {
		t.Fatalf("invalid respect granted=%v err=%v", granted, err)
	}
	if granted, err := service.GrantRespect(context.Background(), 8, 2, "trace:1"); err != nil || !granted || store.respectKey != "trace:1" {
		t.Fatalf("respect granted=%v key=%q err=%v", granted, store.respectKey, err)
	}
	service.Unload(8)
	if _, loaded := service.Wearing(8, "ADM"); loaded {
		t.Fatal("offline snapshot retained")
	}
}

// TestSetEquippedValidatesOwnershipAndRefreshesSnapshot verifies native badge selection.
func TestSetEquippedValidatesOwnershipAndRefreshesSnapshot(t *testing.T) {
	store := &achievementStore{badges: []Badge{{ID: 1, Code: "ADM", Equipped: true, Slot: 1}, {ID: 2, Code: "HC1"}}}
	service := New(store)
	badges, err := service.SetEquipped(context.Background(), 8, []string{"HC1", ""})
	if err != nil || len(store.equipped) != 1 || store.equipped[0] != "HC1" {
		t.Fatalf("badges=%v equipped=%v err=%v", badges, store.equipped, err)
	}
	if wearing, loaded := service.Wearing(8, "HC1"); !wearing || !loaded {
		t.Fatal("new equipped badge missing")
	}
	if wearing, _ := service.Wearing(8, "ADM"); wearing {
		t.Fatal("previous badge remained equipped")
	}
	if _, err = service.SetEquipped(context.Background(), 8, []string{"UNKNOWN"}); !errors.Is(err, ErrBadgeNotOwned) {
		t.Fatalf("ownership error=%v", err)
	}
}

// BenchmarkWearing measures the WIRED equipped-badge hot-path lookup.
func BenchmarkWearing(benchmark *testing.B) {
	service := New(&achievementStore{})
	service.replaceSnapshot(8, []Badge{{Code: "ADM", Equipped: true, Slot: 1}})
	benchmark.ReportAllocs()
	benchmark.ResetTimer()
	for benchmark.Loop() {
		wearing, loaded := service.Wearing(8, "ADM")
		if !wearing || !loaded {
			benchmark.Fatal("equipped badge snapshot missing")
		}
	}
}
