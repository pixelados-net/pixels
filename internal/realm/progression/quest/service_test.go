package quest

import (
	"context"
	"sync"
	"testing"
	"time"

	progressionconfig "github.com/niflaot/pixels/internal/realm/progression/config"
	progressionengine "github.com/niflaot/pixels/internal/realm/progression/engine"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
)

// questCatalogSource returns one fixed quest catalog.
type questCatalogSource struct{ value progressionrecord.Catalog }

// Catalog returns the fixed quest catalog.
func (source questCatalogSource) Catalog(context.Context) (progressionrecord.Catalog, error) {
	return source.value, nil
}

// memoryQuestStore implements the focused quest persistence boundary.
type memoryQuestStore struct {
	progressionrecord.Store
	mutex    sync.Mutex
	active   progressionrecord.PlayerQuestState
	history  map[int64]progressionrecord.PlayerQuestState
	rejected bool
}

// WithinTransaction serializes one in-memory quest transition.
func (store *memoryQuestStore) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	return work(ctx)
}

// ActiveQuest returns the current in-memory quest.
func (store *memoryQuestStore) ActiveQuest(context.Context, int64) (progressionrecord.PlayerQuestState, bool, error) {
	return store.active, store.active.ActiveQuestID != 0, nil
}

// QuestProgress returns durable in-memory quest history.
func (store *memoryQuestStore) QuestProgress(context.Context, int64) ([]progressionrecord.PlayerQuestState, error) {
	values := make([]progressionrecord.PlayerQuestState, 0, len(store.history))
	for _, value := range store.history {
		values = append(values, value)
	}
	return values, nil
}

// ActivateQuest replaces the active quest unless it was completed.
func (store *memoryQuestStore) ActivateQuest(_ context.Context, playerID int64, questID int64) (int64, error) {
	if completed := store.history[questID].CompletedAt; completed != nil {
		return 0, progressionrecord.ErrConflict
	}
	previous := store.active.ActiveQuestID
	store.active = progressionrecord.PlayerQuestState{PlayerID: playerID, ActiveQuestID: questID, AcceptedAt: time.Now()}
	store.history[questID] = store.active
	return previous, nil
}

// IncrementQuest advances the matching active quest up to its goal.
func (store *memoryQuestStore) IncrementQuest(_ context.Context, _ int64, questID int64, amount int64, goal int64) (progressionrecord.PlayerQuestState, error) {
	if store.active.ActiveQuestID != questID {
		return progressionrecord.PlayerQuestState{}, progressionrecord.ErrConflict
	}
	store.active.Progress += amount
	if store.active.Progress > goal {
		store.active.Progress = goal
	}
	store.history[questID] = store.active
	return store.active, nil
}

// TestActivateRejectsSkippedSeriesStage verifies packet ids cannot bypass campaign order.
func TestActivateRejectsSkippedSeriesStage(t *testing.T) {
	campaign := progressionrecord.QuestCampaign{Code: "explore", Enabled: true}
	quests := []progressionrecord.QuestDefinition{
		{ID: 1, CampaignCode: "explore", SeriesNumber: 1, Enabled: true},
		{ID: 2, CampaignCode: "explore", SeriesNumber: 2, Enabled: true},
	}
	service, store := questFixture(t, progressionrecord.Catalog{Campaigns: []progressionrecord.QuestCampaign{campaign}, Quests: quests}, nil)
	if err := service.Activate(context.Background(), 42, 2); err != progressionrecord.ErrConflict {
		t.Fatalf("skipped activation error %v", err)
	}
	if store.active.ActiveQuestID != 0 {
		t.Fatalf("unexpected active quest %d", store.active.ActiveQuestID)
	}
}

// CompleteQuest completes and clears the matching active quest.
func (store *memoryQuestStore) CompleteQuest(_ context.Context, _ int64, questID int64) (bool, error) {
	if store.active.ActiveQuestID != questID {
		return false, nil
	}
	now := time.Now()
	completed := store.active
	completed.CompletedAt = &now
	store.history[questID] = completed
	store.active = progressionrecord.PlayerQuestState{}
	return true, nil
}

// CancelQuest clears and returns the current quest identifier.
func (store *memoryQuestStore) CancelQuest(context.Context, int64) (int64, error) {
	previous := store.active.ActiveQuestID
	store.active = progressionrecord.PlayerQuestState{}
	return previous, nil
}

// DailyQuestRejected reports the in-memory daily rejection flag.
func (store *memoryQuestStore) DailyQuestRejected(context.Context, int64, time.Time) (bool, error) {
	return store.rejected, nil
}

// RejectDailyQuest sets the in-memory daily rejection flag.
func (store *memoryQuestStore) RejectDailyQuest(context.Context, int64, time.Time) error {
	store.rejected = true
	return nil
}

// questFixture builds a loaded quest service.
func questFixture(t testing.TB, catalogValue progressionrecord.Catalog, projector Projector) (*Service, *memoryQuestStore) {
	t.Helper()
	catalog := progressionengine.NewCatalog(questCatalogSource{value: catalogValue})
	if err := catalog.Reload(context.Background()); err != nil {
		t.Fatal(err)
	}
	store := &memoryQuestStore{history: make(map[int64]progressionrecord.PlayerQuestState)}
	return New(progressionconfig.Config{DailyPoolSeed: "qa"}, catalog, store, &questBadges{}, &questCurrencies{}, &questItems{}, projector), store
}

// TestDailyIsDeterministicAndRejectable verifies stable selection and durable rejection.
func TestDailyIsDeterministicAndRejectable(t *testing.T) {
	campaign := progressionrecord.QuestCampaign{Code: "daily", Enabled: true}
	quests := []progressionrecord.QuestDefinition{{ID: 1, CampaignCode: "daily", Daily: true, Easy: true, Enabled: true}, {ID: 2, CampaignCode: "daily", Daily: true, Easy: true, Enabled: true}, {ID: 3, CampaignCode: "daily", Daily: true, Easy: false, Enabled: true}}
	service, _ := questFixture(t, progressionrecord.Catalog{Campaigns: []progressionrecord.QuestCampaign{campaign}, Quests: quests}, nil)
	day := time.Date(2026, 7, 20, 19, 0, 0, 0, time.FixedZone("test", -5*60*60))
	first, found, err := service.Daily(context.Background(), 42, day, true)
	if err != nil || !found {
		t.Fatalf("daily found=%v err=%v", found, err)
	}
	second, found, err := service.Daily(context.Background(), 42, day.Add(2*time.Hour), true)
	if err != nil || !found || first.ID != second.ID {
		t.Fatalf("daily first=%d second=%d found=%v err=%v", first.ID, second.ID, found, err)
	}
	if easy, hard := service.DailyCounts(day); easy != 2 || hard != 1 {
		t.Fatalf("counts easy=%d hard=%d", easy, hard)
	}
	if err = service.RejectDaily(context.Background(), 42, day); err != nil {
		t.Fatal(err)
	}
	if _, found, err = service.Daily(context.Background(), 42, day, true); err != nil || found {
		t.Fatalf("rejected found=%v err=%v", found, err)
	}
}

// TestProgressTriggerRequiresGoalData verifies targeted furniture goals and one-time completion.
func TestProgressTriggerRequiresGoalData(t *testing.T) {
	campaign := progressionrecord.QuestCampaign{Code: "explore", Enabled: true}
	quest := progressionrecord.QuestDefinition{ID: 8, CampaignCode: "explore", TriggerKey: "room.furni.count", GoalData: "1", GoalAmount: 1, RewardKind: "currency", RewardAmount: 5, Enabled: true}
	service, store := questFixture(t, progressionrecord.Catalog{Campaigns: []progressionrecord.QuestCampaign{campaign}, Quests: []progressionrecord.QuestDefinition{quest}}, nil)
	if err := service.Activate(context.Background(), 42, 8); err != nil {
		t.Fatal(err)
	}
	if err := service.ProgressTrigger(context.Background(), 42, "room.furni.count", "2", 1); err != nil {
		t.Fatal(err)
	}
	if store.active.Progress != 0 {
		t.Fatalf("mismatched data progressed to %d", store.active.Progress)
	}
	if err := service.ProgressTrigger(context.Background(), 42, "room.furni.count", "1", 1); err != nil {
		t.Fatal(err)
	}
	if store.active.ActiveQuestID != 0 || store.history[8].CompletedAt == nil {
		t.Fatalf("state %#v history %#v", store.active, store.history[8])
	}
	if err := service.Activate(context.Background(), 42, 8); err != progressionrecord.ErrConflict {
		t.Fatalf("reactivate error %v", err)
	}
}

// TestProgressTriggerCapsDurableProgress verifies large event deltas cannot exceed the quest goal.
func TestProgressTriggerCapsDurableProgress(t *testing.T) {
	campaign := progressionrecord.QuestCampaign{Code: "explore", Enabled: true}
	quest := progressionrecord.QuestDefinition{ID: 8, CampaignCode: "explore", SeriesNumber: 1, TriggerKey: "room.entered", GoalAmount: 3, RewardKind: "currency", Enabled: true}
	service, store := questFixture(t, progressionrecord.Catalog{Campaigns: []progressionrecord.QuestCampaign{campaign}, Quests: []progressionrecord.QuestDefinition{quest}}, nil)
	if err := service.Activate(context.Background(), 42, 8); err != nil {
		t.Fatal(err)
	}
	if err := service.ProgressTrigger(context.Background(), 42, "room.entered", "", 100); err != nil {
		t.Fatal(err)
	}
	if progress := store.history[8].Progress; progress != 3 {
		t.Fatalf("progress=%d want=3", progress)
	}
}

// TestListAdvancesOneOfferPerCampaign verifies chained campaigns expose only their next incomplete stage.
func TestListAdvancesOneOfferPerCampaign(t *testing.T) {
	campaign := progressionrecord.QuestCampaign{Code: "explore", Enabled: true}
	quests := []progressionrecord.QuestDefinition{
		{ID: 1, CampaignCode: "explore", SeriesNumber: 1, Enabled: true},
		{ID: 2, CampaignCode: "explore", SeriesNumber: 2, Enabled: true},
		{ID: 3, CampaignCode: "explore", SeriesNumber: 3, Enabled: true},
	}
	service, store := questFixture(t, progressionrecord.Catalog{Campaigns: []progressionrecord.QuestCampaign{campaign}, Quests: quests}, nil)
	now := time.Now()
	store.history[1] = progressionrecord.PlayerQuestState{ActiveQuestID: 1, CompletedAt: &now}
	offers, progress, err := service.List(context.Background(), 42, false)
	if err != nil || len(offers) != 1 || offers[0].ID != 2 {
		t.Fatalf("offers=%+v err=%v", offers, err)
	}
	if total, completed := service.CampaignCounts(progress, "explore"); total != 3 || completed != 1 {
		t.Fatalf("total=%d completed=%d", total, completed)
	}
	store.active = progressionrecord.PlayerQuestState{PlayerID: 42, ActiveQuestID: 3}
	store.history[3] = store.active
	offers, _, err = service.List(context.Background(), 42, false)
	if err != nil || len(offers) != 1 || offers[0].ID != 3 {
		t.Fatalf("active offer=%+v err=%v", offers, err)
	}
}

// TestDataDoesNotMarkCompletedHistoryActive verifies historical quest identifiers do not look accepted.
func TestDataDoesNotMarkCompletedHistoryActive(t *testing.T) {
	now := time.Now()
	quest := progressionrecord.QuestDefinition{ID: 7}
	value := Data(quest, progressionrecord.PlayerQuestState{ActiveQuestID: 7, CompletedAt: &now}, 1, 1)
	if value.Accepted {
		t.Fatal("completed quest was projected as active")
	}
}
