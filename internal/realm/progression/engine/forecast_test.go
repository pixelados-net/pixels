package engine

import (
	"context"
	"errors"
	"testing"
	"time"

	progressionconfig "github.com/niflaot/pixels/internal/realm/progression/config"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
	"go.uber.org/zap"
)

// capturedQuestProgress records one synchronous quest trigger.
type capturedQuestProgress struct {
	// key stores the received trigger key.
	key string
	// data stores the received goal metadata.
	data string
	// amount stores the received progress delta.
	amount int64
}

// ProgressTrigger records one quest trigger.
func (progress *capturedQuestProgress) ProgressTrigger(_ context.Context, _ int64, key string, data string, amount int64) error {
	progress.key = key
	progress.data = data
	progress.amount = amount
	return nil
}

// TestForecastDetectsHydratedThreshold verifies write-behind can flush a level immediately.
func TestForecastDetectsHydratedThreshold(t *testing.T) {
	service, store, _, _ := progressionFixture(t)
	store.progress.DefinitionID = 7
	store.progress.Progress = 1
	if err := service.HydratePlayer(context.Background(), 42); err != nil {
		t.Fatal(err)
	}
	if !service.add(trigger{playerID: 42, key: "test", amount: 1}) {
		t.Fatal("expected the next cached threshold to request a flush")
	}
	if err := service.FlushPlayer(context.Background(), 42); err != nil {
		t.Fatal(err)
	}
	if store.progress.Level != 1 || store.progress.Progress != 2 {
		t.Fatalf("progress %#v", store.progress)
	}
}

// TestForecastWaitsBelowThreshold verifies ordinary hot-path deltas remain batched.
func TestForecastWaitsBelowThreshold(t *testing.T) {
	service, _, _, _ := progressionFixture(t)
	if err := service.HydratePlayer(context.Background(), 42); err != nil {
		t.Fatal(err)
	}
	if service.add(trigger{playerID: 42, key: "test", amount: 1}) {
		t.Fatal("unexpected threshold flush")
	}
	if service.pendingCount() != 1 {
		t.Fatalf("pending %d", service.pendingCount())
	}
}

// TestWorkerLifecycleFlushesQueuedProgress verifies shutdown drains and commits queued work.
func TestWorkerLifecycleFlushesQueuedProgress(t *testing.T) {
	service, store, _, _ := progressionFixture(t)
	service.config.FlushInterval = time.Hour
	if err := service.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := service.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := service.Progress(context.Background(), 42, "test", 2); err != nil {
		t.Fatal(err)
	}
	if err := service.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := service.Stop(context.Background()); err != nil {
		t.Fatal(err)
	}
	if store.progress.Progress != 2 || store.progress.Level != 1 {
		t.Fatalf("progress %#v", store.progress)
	}
}

// TestFailedFlushRestoresPendingDelta verifies transient persistence failures lose no work.
func TestFailedFlushRestoresPendingDelta(t *testing.T) {
	service, store, _, _ := progressionFixture(t)
	service.add(trigger{playerID: 42, key: "test", amount: 1})
	store.mutationErr = errors.New("write failed")
	if err := service.FlushPlayer(context.Background(), 42); err == nil {
		t.Fatal("expected write failure")
	}
	if service.pendingCount() != 1 {
		t.Fatalf("pending %d", service.pendingCount())
	}
}

// TestDailyBatchDeduplicatesBeforeFlush verifies repeated daily events cannot accumulate in memory.
func TestDailyBatchDeduplicatesBeforeFlush(t *testing.T) {
	service, store, _, _ := progressionFixture(t)
	service.add(trigger{playerID: 42, key: "test", amount: 1, daily: true})
	service.add(trigger{playerID: 42, key: "test", amount: 1, daily: true})
	if err := service.FlushPlayer(context.Background(), 42); err != nil {
		t.Fatal(err)
	}
	if store.progress.Progress != 1 {
		t.Fatalf("daily progress %d", store.progress.Progress)
	}
}

// TestProgressNowDataForwardsQuestMetadata verifies shared triggers retain goal data.
func TestProgressNowDataForwardsQuestMetadata(t *testing.T) {
	definition := progressionrecord.AchievementDefinition{ID: 7, Name: "Test", TriggerKey: "test", Enabled: true, Levels: []progressionrecord.AchievementLevel{{Level: 1, ProgressNeeded: 2}}}
	quest := progressionrecord.QuestDefinition{ID: 8, CampaignCode: "test", TriggerKey: "test", Enabled: true}
	catalog := NewCatalog(catalogSource{catalog: progressionrecord.Catalog{Achievements: []progressionrecord.AchievementDefinition{definition}, Quests: []progressionrecord.QuestDefinition{quest}}})
	if err := catalog.Reload(context.Background()); err != nil {
		t.Fatal(err)
	}
	service := New(progressionconfig.Config{Enabled: true}, catalog, &memoryProgressStore{}, &memoryBadges{}, &memoryCurrencies{}, zap.NewNop())
	quests := &capturedQuestProgress{}
	service.SetQuestProgressor(quests)
	if err := service.ProgressNowData(context.Background(), 42, "test", "furni:44", 2, false); err != nil {
		t.Fatal(err)
	}
	if quests.key != "test" || quests.data != "furni:44" || quests.amount != 2 {
		t.Fatalf("quest progress %#v", quests)
	}
}

// TestProgressNowSkipsQuestLookupForAchievementOnlyTrigger verifies achievement flushes avoid quest persistence.
func TestProgressNowSkipsQuestLookupForAchievementOnlyTrigger(t *testing.T) {
	service, _, _, _ := progressionFixture(t)
	quests := &capturedQuestProgress{}
	service.SetQuestProgressor(quests)
	if err := service.ProgressNow(context.Background(), 42, "test", 1, false); err != nil {
		t.Fatal(err)
	}
	if quests.key != "" {
		t.Fatalf("unexpected quest trigger %#v", quests)
	}
}

// TestForgetPlayerClearsForecastKnowledge verifies disconnect releases player cache entries.
func TestForgetPlayerClearsForecastKnowledge(t *testing.T) {
	service, store, _, _ := progressionFixture(t)
	store.progress.DefinitionID = 7
	store.progress.Progress = 1
	if err := service.HydratePlayer(context.Background(), 42); err != nil {
		t.Fatal(err)
	}
	service.ForgetPlayer(42)
	if service.hydrated[42] {
		t.Fatal("player remained hydrated")
	}
	if _, exists := service.known[playerAchievement{playerID: 42, definitionID: 7}]; exists {
		t.Fatal("player progress remained cached")
	}
}

// TestStartRejectsIncompleteWiring verifies enabled engines fail before starting a worker.
func TestStartRejectsIncompleteWiring(t *testing.T) {
	service := New(progressionconfig.Config{Enabled: true}, nil, nil, nil, nil, zap.NewNop())
	if err := service.Start(context.Background()); err == nil {
		t.Fatal("expected dependency error")
	}
}

// BenchmarkProgressBatch measures the warmed write-behind aggregation path.
func BenchmarkProgressBatch(b *testing.B) {
	service, _, _, _ := progressionFixture(b)
	service.hydrated[42] = true
	value := trigger{playerID: 42, key: "test", amount: 1}
	service.add(value)
	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		service.add(value)
	}
}
