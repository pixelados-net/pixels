package quest

import (
	"context"
	"fmt"
	"time"

	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
)

// reward grants one configured quest reward.
func (service *Service) reward(ctx context.Context, playerID int64, quest progressionrecord.QuestDefinition) error {
	switch quest.RewardKind {
	case "currency":
		_, err := service.currencies.Grant(ctx, currencyservice.GrantParams{PlayerID: playerID, CurrencyType: quest.RewardCurrencyType, Amount: quest.RewardAmount, Reason: "quest reward", ActorKind: currencyservice.ActorSystem})
		return err
	case "badge":
		_, err := service.badges.GrantBadge(ctx, playerID, quest.RewardBadge, "quest")
		return err
	case "item":
		_, err := service.items.Grant(ctx, furnitureservice.GrantParams{DefinitionID: quest.RewardDefinitionID, OwnerPlayerID: playerID, Quantity: int32(quest.RewardAmount), ExtraData: "0"})
		return err
	case "room":
		return nil
	default:
		return fmt.Errorf("%w: reward kind", progressionrecord.ErrInvalid)
	}
}

// ForceComplete activates and completes one available quest through its reward workflow.
func (service *Service) ForceComplete(ctx context.Context, playerID int64, questID int64) error {
	quest, err := service.availableQuest(questID, time.Now())
	if err != nil {
		return err
	}
	var previous int64
	err = service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		var forceErr error
		previous, forceErr = service.store.ActivateQuest(txCtx, playerID, quest.ID)
		if forceErr != nil {
			return forceErr
		}
		if _, forceErr = service.store.IncrementQuest(txCtx, playerID, quest.ID, quest.GoalAmount, quest.GoalAmount); forceErr != nil {
			return forceErr
		}
		completed, forceErr := service.store.CompleteQuest(txCtx, playerID, quest.ID)
		if forceErr != nil {
			return forceErr
		}
		if !completed {
			return progressionrecord.ErrConflict
		}
		return service.reward(txCtx, playerID, quest)
	})
	if err == nil {
		service.metrics.RecordQuestCompleted()
		service.metrics.RecordReward("quest." + quest.RewardKind)
		if service.projector != nil {
			service.projector.Accepted(ctx, playerID, quest, previous)
			service.projector.Completed(ctx, playerID, quest)
			service.refreshList(ctx, playerID)
			if quest.RewardKind == "room" {
				service.projector.RoomReward(ctx, playerID, quest.RewardRoomID)
			}
		}
	}
	return err
}

// availableQuest validates one enabled campaign window.
func (service *Service) availableQuest(questID int64, now time.Time) (progressionrecord.QuestDefinition, error) {
	generation := service.catalog.Current()
	if generation == nil || generation.QuestByID[questID] == nil || !generation.QuestByID[questID].Enabled {
		return progressionrecord.QuestDefinition{}, progressionrecord.ErrNotFound
	}
	quest := *generation.QuestByID[questID]
	campaign := generation.CampaignByCode[quest.CampaignCode]
	if campaign == nil || !campaignAvailable(*campaign, now) {
		return progressionrecord.QuestDefinition{}, progressionrecord.ErrUnavailable
	}
	return quest, nil
}

// campaignAvailable reports whether one campaign is currently open.
func campaignAvailable(campaign progressionrecord.QuestCampaign, now time.Time) bool {
	if !campaign.Enabled || campaign.StartsAt != nil && now.Before(*campaign.StartsAt) {
		return false
	}
	return campaign.EndsAt == nil || now.Before(*campaign.EndsAt)
}
