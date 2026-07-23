package quest

import (
	"context"
	"testing"
	"time"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
)

// questBadges records badge reward calls.
type questBadges struct{ count int }

// GrantBadge records one badge reward.
func (badges *questBadges) GrantBadge(context.Context, int64, string, string) (bool, error) {
	badges.count++
	return true, nil
}

// questCurrencies records wallet reward calls.
type questCurrencies struct{ count int }

// Grant records one wallet reward.
func (currencies *questCurrencies) Grant(context.Context, currencyservice.GrantParams) (int64, error) {
	currencies.count++
	return int64(currencies.count), nil
}

// questItems records furniture reward calls.
type questItems struct{ count int }

// Grant records one furniture reward.
func (items *questItems) Grant(context.Context, furnitureservice.GrantParams) ([]furnituremodel.Item, error) {
	items.count++
	return []furnituremodel.Item{{}}, nil
}

// questProjector records committed quest projections.
type questProjector struct {
	completed int
	listed    int
	roomID    int64
}

// Accepted records one activation.
func (*questProjector) Accepted(context.Context, int64, progressionrecord.QuestDefinition, int64) {}

// Progressed records one progress update.
func (*questProjector) Progressed(context.Context, int64, progressionrecord.QuestDefinition, progressionrecord.PlayerQuestState) {
}

// Completed records one completion.
func (projector *questProjector) Completed(context.Context, int64, progressionrecord.QuestDefinition) {
	projector.completed++
}

// Listed records one refreshed quest list.
func (projector *questProjector) Listed(context.Context, int64, []progressionrecord.QuestDefinition, map[int64]progressionrecord.PlayerQuestState) {
	projector.listed++
}

// Cancelled records one cancellation.
func (*questProjector) Cancelled(context.Context, int64, bool) {}

// RoomReward records one room forwarding reward.
func (projector *questProjector) RoomReward(_ context.Context, _ int64, roomID int64) {
	projector.roomID = roomID
}

// TestRewardKinds verifies every supported quest reward workflow.
func TestRewardKinds(t *testing.T) {
	for _, kind := range []string{"currency", "badge", "item", "room"} {
		t.Run(kind, func(t *testing.T) {
			campaign := progressionrecord.QuestCampaign{Code: "qa", Enabled: true}
			quest := progressionrecord.QuestDefinition{ID: 1, CampaignCode: "qa", GoalAmount: 1, RewardKind: kind, RewardAmount: 1, RewardBadge: "QA", RewardDefinitionID: 1, RewardRoomID: 130, Enabled: true}
			projector := &questProjector{}
			service, _ := questFixture(t, progressionrecord.Catalog{Campaigns: []progressionrecord.QuestCampaign{campaign}, Quests: []progressionrecord.QuestDefinition{quest}}, projector)
			badges, currencies, items := &questBadges{}, &questCurrencies{}, &questItems{}
			service.badges, service.currencies, service.items = badges, currencies, items
			if err := service.ForceComplete(context.Background(), 42, 1); err != nil {
				t.Fatal(err)
			}
			if projector.completed != 1 || projector.listed != 1 {
				t.Fatalf("completion projections completed=%d listed=%d", projector.completed, projector.listed)
			}
			counts := map[string]int{"currency": currencies.count, "badge": badges.count, "item": items.count}
			if kind != "room" && counts[kind] != 1 {
				t.Fatalf("reward counts %#v", counts)
			}
			if kind == "room" && projector.roomID != 130 {
				t.Fatalf("room reward %d", projector.roomID)
			}
		})
	}
}

// TestSeasonalWindowRejectsActivation verifies closed campaigns never activate.
func TestSeasonalWindowRejectsActivation(t *testing.T) {
	start, end := timePair()
	campaign := progressionrecord.QuestCampaign{Code: "closed", Seasonal: true, StartsAt: &start, EndsAt: &end, Enabled: true}
	quest := progressionrecord.QuestDefinition{ID: 1, CampaignCode: "closed", GoalAmount: 1, RewardKind: "currency", Enabled: true}
	service, _ := questFixture(t, progressionrecord.Catalog{Campaigns: []progressionrecord.QuestCampaign{campaign}, Quests: []progressionrecord.QuestDefinition{quest}}, nil)
	if err := service.Activate(context.Background(), 42, 1); err != progressionrecord.ErrUnavailable {
		t.Fatalf("activation error %v", err)
	}
}

// timePair returns one expired campaign window.
func timePair() (time.Time, time.Time) {
	return time.Now().Add(-2 * time.Hour), time.Now().Add(-time.Hour)
}
