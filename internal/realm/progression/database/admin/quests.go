package admin

import (
	"context"

	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
)

// CreateCampaign inserts one quest campaign.
func (repository *Repository) CreateCampaign(ctx context.Context, value progressionrecord.QuestCampaign) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `insert into quest_campaigns(code,seasonal,starts_at,ends_at,timing_code,enabled) values($1,$2,$3,$4,$5,$6)`, value.Code, value.Seasonal, value.StartsAt, value.EndsAt, value.TimingCode, value.Enabled)
	return err
}

// UpdateCampaign replaces one quest campaign.
func (repository *Repository) UpdateCampaign(ctx context.Context, value progressionrecord.QuestCampaign) (bool, error) {
	result, err := repository.executorFor(ctx).Exec(ctx, `update quest_campaigns set seasonal=$2,starts_at=$3,ends_at=$4,timing_code=$5,enabled=$6 where code=$1`, value.Code, value.Seasonal, value.StartsAt, value.EndsAt, value.TimingCode, value.Enabled)
	return result.RowsAffected() > 0, err
}

// DisableCampaign soft-disables one quest campaign.
func (repository *Repository) DisableCampaign(ctx context.Context, code string) (bool, error) {
	result, err := repository.executorFor(ctx).Exec(ctx, `update quest_campaigns set enabled=false where code=$1 and enabled`, code)
	return result.RowsAffected() > 0, err
}

// CreateQuest inserts one quest definition.
func (repository *Repository) CreateQuest(ctx context.Context, value progressionrecord.QuestDefinition) (progressionrecord.QuestDefinition, error) {
	err := repository.executorFor(ctx).QueryRow(ctx, `insert into quest_definitions(campaign_code,series_number,name,localization_code,trigger_key,goal_amount,goal_data,reward_kind,reward_currency_type,reward_amount,reward_badge,reward_definition_id,reward_room_id,daily,easy,sort_order,enabled) values($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,nullif($12,0),nullif($13,0),$14,$15,$16,$17) returning id,version`, value.CampaignCode, value.SeriesNumber, value.Name, value.LocalizationCode, value.TriggerKey, value.GoalAmount, value.GoalData, value.RewardKind, value.RewardCurrencyType, value.RewardAmount, value.RewardBadge, value.RewardDefinitionID, value.RewardRoomID, value.Daily, value.Easy, value.SortOrder, value.Enabled).Scan(&value.ID, &value.Version)
	return value, err
}

// UpdateQuest replaces mutable fields under optimistic locking.
func (repository *Repository) UpdateQuest(ctx context.Context, value progressionrecord.QuestDefinition, version int64) (progressionrecord.QuestDefinition, error) {
	err := repository.executorFor(ctx).QueryRow(ctx, `update quest_definitions set campaign_code=$2,series_number=$3,name=$4,localization_code=$5,trigger_key=$6,goal_amount=$7,goal_data=$8,reward_kind=$9,reward_currency_type=$10,reward_amount=$11,reward_badge=$12,reward_definition_id=nullif($13,0),reward_room_id=nullif($14,0),daily=$15,easy=$16,sort_order=$17,enabled=$18,version=version+1 where id=$1 and version=$19 returning version`, value.ID, value.CampaignCode, value.SeriesNumber, value.Name, value.LocalizationCode, value.TriggerKey, value.GoalAmount, value.GoalData, value.RewardKind, value.RewardCurrencyType, value.RewardAmount, value.RewardBadge, value.RewardDefinitionID, value.RewardRoomID, value.Daily, value.Easy, value.SortOrder, value.Enabled, version).Scan(&value.Version)
	if noRows(err) {
		err = repository.missingOrConflict(ctx, "quest_definitions", "id", value.ID)
	}
	return value, err
}

// DisableQuest soft-disables one quest definition.
func (repository *Repository) DisableQuest(ctx context.Context, id int64) (bool, error) {
	result, err := repository.executorFor(ctx).Exec(ctx, `update quest_definitions set enabled=false,version=version+1 where id=$1 and enabled`, id)
	return result.RowsAffected() > 0, err
}
