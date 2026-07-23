package engine

import (
	"context"
	"fmt"
	"sync"
	"testing"

	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	progressionconfig "github.com/niflaot/pixels/internal/realm/progression/config"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
	"go.uber.org/zap"
)

// memoryProgressStore implements the progression mutation boundary for focused tests.
type memoryProgressStore struct {
	progressionrecord.Store
	mutex       sync.Mutex
	progress    progressionrecord.PlayerAchievement
	score       int32
	mutationErr error
}

// PlayerAchievements returns the current in-memory progress when present.
func (store *memoryProgressStore) PlayerAchievements(context.Context, int64) ([]progressionrecord.PlayerAchievement, error) {
	if store.progress.DefinitionID == 0 {
		return nil, nil
	}
	return []progressionrecord.PlayerAchievement{store.progress}, nil
}

// WithinTransaction serializes one in-memory mutation.
func (store *memoryProgressStore) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	return work(ctx)
}

// MutateProgress applies one monotonic in-memory delta.
func (store *memoryProgressStore) MutateProgress(_ context.Context, playerID int64, definition progressionrecord.AchievementDefinition, amount int64, _ bool) (progressionrecord.ProgressMutation, error) {
	if store.mutationErr != nil {
		return progressionrecord.ProgressMutation{}, store.mutationErr
	}
	before := store.progress
	after := before
	after.PlayerID, after.DefinitionID = playerID, definition.ID
	after.Progress += amount
	if maximum := definition.Levels[len(definition.Levels)-1].ProgressNeeded; after.Progress > maximum {
		after.Progress = maximum
	}
	after.Level = levelForProgress(definition.Levels, after.Progress)
	crossed := make([]progressionrecord.AchievementLevel, 0, after.Level-before.Level)
	for _, level := range definition.Levels {
		if level.Level > before.Level && level.Level <= after.Level {
			crossed = append(crossed, level)
		}
	}
	store.progress = after
	return progressionrecord.ProgressMutation{Before: before, After: after, Crossed: crossed}, nil
}

// SetProgress replaces one exact in-memory state.
func (store *memoryProgressStore) SetProgress(_ context.Context, playerID int64, definition progressionrecord.AchievementDefinition, progress int64, level int32, pay bool) (progressionrecord.ProgressMutation, error) {
	before := store.progress
	after := progressionrecord.PlayerAchievement{PlayerID: playerID, DefinitionID: definition.ID, Progress: progress, Level: level}
	crossed := []progressionrecord.AchievementLevel(nil)
	if pay {
		for _, candidate := range definition.Levels {
			if candidate.Level > before.Level && candidate.Level <= level {
				crossed = append(crossed, candidate)
			}
		}
	}
	store.progress = after
	return progressionrecord.ProgressMutation{Before: before, After: after, Crossed: crossed}, nil
}

// AddScore applies one in-memory score delta.
func (store *memoryProgressStore) AddScore(_ context.Context, _ int64, amount int32) (int32, error) {
	store.score += amount
	return store.score, nil
}

// ResetProgress clears the in-memory state.
func (store *memoryProgressStore) ResetProgress(context.Context, int64, int64) (bool, error) {
	store.progress = progressionrecord.PlayerAchievement{}
	return true, nil
}

// memoryBadges records the single badge owned by one achievement group.
type memoryBadges struct {
	code string
}

// GrantBadge records one badge idempotently.
func (badges *memoryBadges) GrantBadge(_ context.Context, _ int64, code string, _ string) (bool, error) {
	if badges.code == code {
		return false, nil
	}
	badges.code = code
	return true, nil
}

// ReplaceBadge advances the badge only when the expected old code exists.
func (badges *memoryBadges) ReplaceBadge(_ context.Context, _ int64, oldCode string, newCode string, _ string) (bool, error) {
	if badges.code != oldCode {
		return false, nil
	}
	badges.code = newCode
	return true, nil
}

// RemoveBadge clears the matching badge.
func (badges *memoryBadges) RemoveBadge(_ context.Context, _ int64, code string) (bool, error) {
	if badges.code != code {
		return false, nil
	}
	badges.code = ""
	return true, nil
}

// memoryCurrencies records committed reward grants.
type memoryCurrencies struct {
	mutex  sync.Mutex
	grants []currencyservice.GrantParams
}

// Grant records one currency reward.
func (currencies *memoryCurrencies) Grant(_ context.Context, params currencyservice.GrantParams) (int64, error) {
	currencies.mutex.Lock()
	defer currencies.mutex.Unlock()
	currencies.grants = append(currencies.grants, params)
	return int64(len(currencies.grants)), nil
}

// progressionFixture builds one three-level cached progression service.
func progressionFixture(t testing.TB) (*Service, *memoryProgressStore, *memoryBadges, *memoryCurrencies) {
	t.Helper()
	levels := []progressionrecord.AchievementLevel{{Level: 1, ProgressNeeded: 2, RewardAmount: 1, ScorePoints: 5}, {Level: 2, ProgressNeeded: 5, RewardAmount: 2, ScorePoints: 10}, {Level: 3, ProgressNeeded: 9, RewardAmount: 3, ScorePoints: 15}}
	definition := progressionrecord.AchievementDefinition{ID: 7, Name: "Test", TriggerKey: "test", Enabled: true, Levels: levels}
	catalog := NewCatalog(catalogSource{catalog: progressionrecord.Catalog{Achievements: []progressionrecord.AchievementDefinition{definition}}})
	if err := catalog.Reload(context.Background()); err != nil {
		t.Fatal(err)
	}
	store, badges, currencies := &memoryProgressStore{}, &memoryBadges{}, &memoryCurrencies{}
	service := New(progressionconfig.Config{Enabled: true}, catalog, store, badges, currencies, zap.NewNop())
	return service, store, badges, currencies
}

// TestProgressDefinitionCrossesEveryLevel verifies multi-level rewards and badge replacement.
func TestProgressDefinitionCrossesEveryLevel(t *testing.T) {
	service, store, badges, currencies := progressionFixture(t)
	if err := service.ProgressDefinition(context.Background(), 42, 7, 9); err != nil {
		t.Fatal(err)
	}
	if store.progress.Level != 3 || store.score != 30 || badges.code != "ACH_Test3" || len(currencies.grants) != 3 {
		t.Fatalf("progress=%#v score=%d badge=%q grants=%d", store.progress, store.score, badges.code, len(currencies.grants))
	}
}

// TestConcurrentProgressPaysEachLevelOnce verifies serialized level rewards under concurrent sessions.
func TestConcurrentProgressPaysEachLevelOnce(t *testing.T) {
	service, store, badges, currencies := progressionFixture(t)
	var wait sync.WaitGroup
	for range 8 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			if err := service.ProgressDefinition(context.Background(), 42, 7, 2); err != nil {
				t.Errorf("progress: %v", err)
			}
		}()
	}
	wait.Wait()
	if store.progress.Level != 3 || badges.code != "ACH_Test3" || len(currencies.grants) != 3 {
		t.Fatalf("progress=%#v badge=%q grants=%d", store.progress, badges.code, len(currencies.grants))
	}
}

// TestSetLevelWithoutRewardsAlignsBadge verifies corrective overrides never repay rewards.
func TestSetLevelWithoutRewardsAlignsBadge(t *testing.T) {
	service, store, badges, currencies := progressionFixture(t)
	if err := service.SetLevel(context.Background(), 42, 7, 2, false); err != nil {
		t.Fatal(err)
	}
	if store.progress.Level != 2 || badges.code != "ACH_Test2" || len(currencies.grants) != 0 {
		t.Fatal(fmt.Sprintf("progress=%#v badge=%q grants=%d", store.progress, badges.code, len(currencies.grants)))
	}
}
