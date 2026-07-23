package quest

import (
	"time"

	progressionengine "github.com/niflaot/pixels/internal/realm/progression/engine"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
	questdata "github.com/niflaot/pixels/networking/outbound/progression/quest/data"
)

// Data maps one quest and player state to Nitro's exact wire record.
func Data(quest progressionrecord.QuestDefinition, state progressionrecord.PlayerQuestState, campaignCount int32, completed int32) questdata.Quest {
	return questdata.Quest{
		CampaignCode: quest.CampaignCode, CompletedInCampaign: completed, CampaignCount: campaignCount,
		RewardCurrencyType: quest.RewardCurrencyType, ID: progressionengine.Clamp(quest.ID), Accepted: state.ActiveQuestID == quest.ID && state.CompletedAt == nil,
		Type: quest.TriggerKey, RewardAmount: progressionengine.Clamp(quest.RewardAmount), LocalizationCode: quest.LocalizationCode,
		CompletedSteps: progressionengine.Clamp(state.Progress), TotalSteps: progressionengine.Clamp(quest.GoalAmount), SortOrder: quest.SortOrder,
		CatalogPage: quest.GoalData, ChainCode: quest.Name, Easy: quest.Easy,
	}
}

// CampaignCounts returns the full enabled series size and completed count.
func (service *Service) CampaignCounts(history map[int64]progressionrecord.PlayerQuestState, code string) (int32, int32) {
	generation := service.catalog.Current()
	if generation == nil {
		return 0, 0
	}
	return CampaignCounts(generation.Catalog.Quests, history, code)
}

// CampaignCounts derives the total and completed rows for one quest campaign.
func CampaignCounts(catalog []progressionrecord.QuestDefinition, history map[int64]progressionrecord.PlayerQuestState, code string) (int32, int32) {
	var total int32
	var completed int32
	for _, quest := range catalog {
		if quest.CampaignCode != code || !quest.Enabled {
			continue
		}
		total++
		if state, found := history[quest.ID]; found && state.CompletedAt != nil {
			completed++
		}
	}
	return total, completed
}

// nextCampaignQuests selects one active, next incomplete, or completed terminal quest per campaign.
func nextCampaignQuests(catalog progressionrecord.Catalog, history map[int64]progressionrecord.PlayerQuestState, seasonal bool, now time.Time) []progressionrecord.QuestDefinition {
	values := make([]progressionrecord.QuestDefinition, 0, len(catalog.Campaigns))
	for _, campaign := range catalog.Campaigns {
		if campaign.Seasonal != seasonal || !campaignAvailable(campaign, now) {
			continue
		}
		if quest, found := campaignQuest(catalog.Quests, history, campaign.Code); found {
			values = append(values, quest)
		}
	}
	return values
}

// campaignQuest selects one campaign's current client-facing series entry.
func campaignQuest(quests []progressionrecord.QuestDefinition, history map[int64]progressionrecord.PlayerQuestState, code string) (progressionrecord.QuestDefinition, bool) {
	var next progressionrecord.QuestDefinition
	var terminal progressionrecord.QuestDefinition
	for _, quest := range quests {
		if !quest.Enabled || quest.CampaignCode != code {
			continue
		}
		if terminal.ID == 0 || quest.SeriesNumber > terminal.SeriesNumber {
			terminal = quest
		}
		state, found := history[quest.ID]
		if found && state.ActiveQuestID == quest.ID && state.CompletedAt == nil {
			return quest, true
		}
		if (!found || state.CompletedAt == nil) && (next.ID == 0 || quest.SeriesNumber < next.SeriesNumber) {
			next = quest
		}
	}
	if next.ID != 0 {
		return next, true
	}
	return terminal, terminal.ID != 0
}
