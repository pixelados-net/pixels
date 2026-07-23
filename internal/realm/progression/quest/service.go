// Package quest implements campaigns, active quests, and rewards.
package quest

import (
	"context"
	"time"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	progressionconfig "github.com/niflaot/pixels/internal/realm/progression/config"
	progressionengine "github.com/niflaot/pixels/internal/realm/progression/engine"
	progressionobservability "github.com/niflaot/pixels/internal/realm/progression/observability"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
)

// BadgeGranter grants durable quest badge rewards.
type BadgeGranter interface {
	// GrantBadge grants one badge idempotently.
	GrantBadge(context.Context, int64, string, string) (bool, error)
}

// ItemGranter grants furniture inventory rewards.
type ItemGranter interface {
	// Grant creates inventory furniture.
	Grant(context.Context, furnitureservice.GrantParams) ([]furnituremodel.Item, error)
}

// Projector publishes committed quest transitions.
type Projector interface {
	// Accepted publishes one activation and optional cancellation.
	Accepted(context.Context, int64, progressionrecord.QuestDefinition, int64)
	// Progressed publishes one current quest update.
	Progressed(context.Context, int64, progressionrecord.QuestDefinition, progressionrecord.PlayerQuestState)
	// Completed publishes one completed quest and next offer.
	Completed(context.Context, int64, progressionrecord.QuestDefinition)
	// Listed publishes the refreshed available quest list.
	Listed(context.Context, int64, []progressionrecord.QuestDefinition, map[int64]progressionrecord.PlayerQuestState)
	// Cancelled publishes one quest cancellation.
	Cancelled(context.Context, int64, bool)
	// RoomReward forwards one online player to a configured reward room.
	RoomReward(context.Context, int64, int64)
}

// Service owns quest lifecycle and deterministic daily selection.
type Service struct {
	// config stores deterministic selection policy.
	config progressionconfig.Config
	// catalog owns immutable definitions.
	catalog *progressionengine.Catalog
	// store persists quest state.
	store progressionrecord.Store
	// badges owns badge rewards.
	badges BadgeGranter
	// currencies owns wallet rewards.
	currencies currencyservice.Granter
	// items owns furniture rewards.
	items ItemGranter
	// projector publishes client changes.
	projector Projector
	// metrics stores process-wide progression telemetry.
	metrics *progressionobservability.Metrics
}

// SetMetrics attaches process-wide telemetry before serving quests.
func (service *Service) SetMetrics(metrics *progressionobservability.Metrics) {
	service.metrics = metrics
}

// New creates a quest service.
func New(config progressionconfig.Config, catalog *progressionengine.Catalog, store progressionrecord.Store, badges BadgeGranter, currencies currencyservice.Granter, items ItemGranter, projectors ...Projector) *Service {
	service := &Service{config: config, catalog: catalog, store: store, badges: badges, currencies: currencies, items: items}
	if len(projectors) > 0 {
		service.projector = projectors[0]
	}
	return service
}

// Timing returns the active campaign timing code and its closing instant.
func (service *Service) Timing(now time.Time) (string, string) {
	generation := service.catalog.Current()
	if generation == nil {
		return "", ""
	}
	for _, campaign := range generation.Catalog.Campaigns {
		if campaign.TimingCode == "" || !campaignAvailable(campaign, now) {
			continue
		}
		until := ""
		if campaign.EndsAt != nil {
			until = campaign.EndsAt.UTC().Format(time.RFC3339)
		}
		return campaign.TimingCode, until
	}
	return "", ""
}

// List returns available quests with their current player progress.
func (service *Service) List(ctx context.Context, playerID int64, seasonalOnly bool) ([]progressionrecord.QuestDefinition, map[int64]progressionrecord.PlayerQuestState, error) {
	generation := service.catalog.Current()
	if generation == nil {
		return nil, nil, nil
	}
	history, err := service.store.QuestProgress(ctx, playerID)
	if err != nil {
		return nil, nil, err
	}
	progress := make(map[int64]progressionrecord.PlayerQuestState, len(history))
	for _, value := range history {
		progress[value.ActiveQuestID] = value
	}
	now := time.Now()
	return nextCampaignQuests(generation.Catalog, progress, seasonalOnly, now), progress, nil
}

// Active returns the current quest and progress.
func (service *Service) Active(ctx context.Context, playerID int64) (progressionrecord.QuestDefinition, progressionrecord.PlayerQuestState, bool, error) {
	state, found, err := service.store.ActiveQuest(ctx, playerID)
	if err != nil || !found {
		return progressionrecord.QuestDefinition{}, state, false, err
	}
	generation := service.catalog.Current()
	if generation == nil || generation.QuestByID[state.ActiveQuestID] == nil {
		return progressionrecord.QuestDefinition{}, state, false, nil
	}
	return *generation.QuestByID[state.ActiveQuestID], state, true, nil
}

// Activate replaces one active quest explicitly.
func (service *Service) Activate(ctx context.Context, playerID int64, questID int64) error {
	quest, err := service.availableQuest(questID, time.Now())
	if err != nil {
		return err
	}
	if err = service.validateOffer(ctx, playerID, quest); err != nil {
		return err
	}
	var previous int64
	err = service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		var activateErr error
		previous, activateErr = service.store.ActivateQuest(txCtx, playerID, quest.ID)
		return activateErr
	})
	if err == nil && service.projector != nil {
		service.projector.Accepted(ctx, playerID, quest, previous)
	}
	return err
}

// Cancel clears one active quest.
func (service *Service) Cancel(ctx context.Context, playerID int64, expired bool) error {
	questID, err := service.store.CancelQuest(ctx, playerID)
	if err == nil && questID != 0 && service.projector != nil {
		service.projector.Cancelled(ctx, playerID, expired)
	}
	return err
}

// ProgressTrigger advances a matching active quest and completes it atomically.
func (service *Service) ProgressTrigger(ctx context.Context, playerID int64, key string, data string, amount int64) error {
	if amount <= 0 {
		return nil
	}
	quest, _, found, err := service.Active(ctx, playerID)
	if err != nil || !found || quest.TriggerKey != key || quest.GoalData != "" && quest.GoalData != data {
		return err
	}
	campaign := service.catalog.Current().CampaignByCode[quest.CampaignCode]
	if campaign == nil || !campaignAvailable(*campaign, time.Now()) {
		return nil
	}
	var state progressionrecord.PlayerQuestState
	completed := false
	err = service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		var incrementErr error
		state, incrementErr = service.store.IncrementQuest(txCtx, playerID, quest.ID, amount, quest.GoalAmount)
		if incrementErr != nil || state.Progress < quest.GoalAmount {
			return incrementErr
		}
		completed, incrementErr = service.store.CompleteQuest(txCtx, playerID, quest.ID)
		if incrementErr != nil || !completed {
			return incrementErr
		}
		return service.reward(txCtx, playerID, quest)
	})
	if err != nil {
		return err
	}
	if service.projector != nil {
		if completed {
			service.projector.Completed(ctx, playerID, quest)
			service.refreshList(ctx, playerID)
			if quest.RewardKind == "room" {
				service.projector.RoomReward(ctx, playerID, quest.RewardRoomID)
			}
		} else {
			service.projector.Progressed(ctx, playerID, quest, state)
		}
	}
	if completed {
		service.metrics.RecordQuestCompleted()
		service.metrics.RecordReward("quest." + quest.RewardKind)
	}
	return nil
}

// validateOffer prevents clients from skipping unfinished campaign stages.
func (service *Service) validateOffer(ctx context.Context, playerID int64, quest progressionrecord.QuestDefinition) error {
	history, err := service.store.QuestProgress(ctx, playerID)
	if err != nil {
		return err
	}
	progress := make(map[int64]progressionrecord.PlayerQuestState, len(history))
	for _, state := range history {
		progress[state.ActiveQuestID] = state
	}
	offer, found := campaignQuest(service.catalog.Current().Catalog.Quests, progress, quest.CampaignCode)
	if !found || offer.ID != quest.ID {
		return progressionrecord.ErrConflict
	}
	return nil
}

// refreshList publishes post-completion offers without making committed rewards fail on projection reads.
func (service *Service) refreshList(ctx context.Context, playerID int64) {
	quests, progress, err := service.List(ctx, playerID, false)
	if err == nil && service.projector != nil {
		service.projector.Listed(ctx, playerID, quests, progress)
	}
}
