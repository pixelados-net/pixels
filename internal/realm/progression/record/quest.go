package record

import "time"

// QuestCampaign describes one quest family and availability window.
type QuestCampaign struct {
	// Code identifies the campaign.
	Code string
	// Seasonal reports whether availability is windowed.
	Seasonal bool
	// StartsAt stores the optional opening instant.
	StartsAt *time.Time
	// EndsAt stores the optional closing instant.
	EndsAt *time.Time
	// TimingCode stores the Nitro timing localization code.
	TimingCode string
	// Enabled controls client and engine visibility.
	Enabled bool
}

// QuestDefinition describes one data-driven quest.
type QuestDefinition struct {
	// ID identifies the quest.
	ID int64
	// CampaignCode identifies the parent campaign.
	CampaignCode string
	// SeriesNumber orders quests inside the campaign.
	SeriesNumber int32
	// Name stores the stable quest name.
	Name string
	// LocalizationCode stores the client localization suffix.
	LocalizationCode string
	// TriggerKey identifies the shared progression signal.
	TriggerKey string
	// GoalAmount stores the completion threshold.
	GoalAmount int64
	// GoalData stores optional trigger-specific data.
	GoalData string
	// RewardKind selects currency, badge, item, or room behavior.
	RewardKind string
	// RewardCurrencyType identifies a wallet reward.
	RewardCurrencyType int32
	// RewardAmount stores the reward quantity.
	RewardAmount int64
	// RewardBadge stores an optional badge reward.
	RewardBadge string
	// RewardDefinitionID stores an optional furniture reward.
	RewardDefinitionID int64
	// RewardRoomID stores an optional destination room.
	RewardRoomID int64
	// Daily reports membership in the daily pool.
	Daily bool
	// Easy distinguishes the daily difficulty counter.
	Easy bool
	// SortOrder stores client ordering.
	SortOrder int32
	// Enabled controls availability.
	Enabled bool
	// Version stores optimistic administration state.
	Version int64
}

// PlayerQuestState stores the single active quest invariant.
type PlayerQuestState struct {
	// PlayerID identifies the player.
	PlayerID int64
	// ActiveQuestID identifies the current quest or zero.
	ActiveQuestID int64
	// AcceptedAt stores the activation instant.
	AcceptedAt time.Time
	// Progress stores active quest progress.
	Progress int64
	// CompletedAt stores the optional completion instant.
	CompletedAt *time.Time
}
