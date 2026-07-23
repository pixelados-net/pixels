package request

import (
	"strings"
	"time"

	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
)

// Campaign creates or replaces one quest campaign.
type Campaign struct {
	Audit
	// Code identifies a campaign during creation.
	Code string `json:"code"`
	// Seasonal reports whether availability is windowed.
	Seasonal bool `json:"seasonal"`
	// StartsAt stores the optional opening instant.
	StartsAt *time.Time `json:"startsAt"`
	// EndsAt stores the optional closing instant.
	EndsAt *time.Time `json:"endsAt"`
	// TimingCode stores the Nitro localization code.
	TimingCode string `json:"timingCode"`
	// Enabled optionally controls availability and defaults true.
	Enabled *bool `json:"enabled"`
}

// Value maps one campaign request to a domain record.
func (request Campaign) Value(code string) progressionrecord.QuestCampaign {
	enabled := true
	if request.Enabled != nil {
		enabled = *request.Enabled
	}
	return progressionrecord.QuestCampaign{Code: strings.TrimSpace(code), Seasonal: request.Seasonal, StartsAt: request.StartsAt, EndsAt: request.EndsAt, TimingCode: strings.TrimSpace(request.TimingCode), Enabled: enabled}
}

// Quest creates or replaces one quest definition.
type Quest struct {
	Audit
	// Version stores the expected version during updates.
	Version int64 `json:"version"`
	// CampaignCode identifies the parent campaign.
	CampaignCode string `json:"campaignCode"`
	// SeriesNumber orders quests inside the campaign.
	SeriesNumber int32 `json:"seriesNumber"`
	// Name stores the stable quest name.
	Name string `json:"name"`
	// LocalizationCode stores the client localization suffix.
	LocalizationCode string `json:"localizationCode"`
	// TriggerKey identifies the progression signal.
	TriggerKey string `json:"triggerKey"`
	// GoalAmount stores the completion threshold.
	GoalAmount int64 `json:"goalAmount"`
	// GoalData stores optional trigger-specific data.
	GoalData string `json:"goalData"`
	// RewardKind selects currency, badge, item, or room behavior.
	RewardKind string `json:"rewardKind"`
	// RewardCurrencyType identifies a wallet reward.
	RewardCurrencyType int32 `json:"rewardCurrencyType"`
	// RewardAmount stores the reward quantity.
	RewardAmount int64 `json:"rewardAmount"`
	// RewardBadge stores an optional badge reward.
	RewardBadge string `json:"rewardBadge"`
	// RewardDefinitionID stores an optional furniture reward.
	RewardDefinitionID int64 `json:"rewardDefinitionId"`
	// RewardRoomID stores an optional destination room.
	RewardRoomID int64 `json:"rewardRoomId"`
	// Daily reports membership in the daily pool.
	Daily bool `json:"daily"`
	// Easy distinguishes daily difficulty.
	Easy bool `json:"easy"`
	// SortOrder stores client ordering.
	SortOrder int32 `json:"sortOrder"`
	// Enabled optionally controls availability and defaults true.
	Enabled *bool `json:"enabled"`
}

// Value maps one full quest request to a domain record.
func (request Quest) Value(id int64) progressionrecord.QuestDefinition {
	enabled := true
	if request.Enabled != nil {
		enabled = *request.Enabled
	}
	return progressionrecord.QuestDefinition{ID: id, CampaignCode: strings.TrimSpace(request.CampaignCode), SeriesNumber: request.SeriesNumber, Name: strings.TrimSpace(request.Name), LocalizationCode: strings.TrimSpace(request.LocalizationCode), TriggerKey: strings.TrimSpace(request.TriggerKey), GoalAmount: request.GoalAmount, GoalData: strings.TrimSpace(request.GoalData), RewardKind: strings.TrimSpace(request.RewardKind), RewardCurrencyType: request.RewardCurrencyType, RewardAmount: request.RewardAmount, RewardBadge: strings.TrimSpace(request.RewardBadge), RewardDefinitionID: request.RewardDefinitionID, RewardRoomID: request.RewardRoomID, Daily: request.Daily, Easy: request.Easy, SortOrder: request.SortOrder, Enabled: enabled, Version: request.Version}
}

// ValidCampaign validates one campaign window.
func ValidCampaign(value progressionrecord.QuestCampaign) bool {
	return value.Code != "" && (value.StartsAt == nil || value.EndsAt == nil || value.StartsAt.Before(*value.EndsAt))
}

// ValidQuest validates one quest definition.
func ValidQuest(value progressionrecord.QuestDefinition) bool {
	validReward := value.RewardKind == "currency" || value.RewardKind == "badge" || value.RewardKind == "item" || value.RewardKind == "room"
	return value.CampaignCode != "" && value.SeriesNumber > 0 && value.Name != "" && value.LocalizationCode != "" && value.TriggerKey != "" && value.GoalAmount > 0 && value.RewardAmount >= 0 && validReward
}
