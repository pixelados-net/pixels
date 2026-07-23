package admin

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
)

// CreateAchievement inserts one achievement definition.
func (repository *Repository) CreateAchievement(ctx context.Context, value progressionrecord.AchievementDefinition) (progressionrecord.AchievementDefinition, error) {
	err := repository.executorFor(ctx).QueryRow(ctx, `insert into achievement_definitions(name,category,subcategory,trigger_key,visible,enabled) values($1,$2,$3,$4,$5,$6) returning id,version`, value.Name, value.Category, value.Subcategory, value.TriggerKey, value.Visible, value.Enabled).Scan(&value.ID, &value.Version)
	return value, err
}

// UpdateAchievement replaces mutable fields under optimistic locking.
func (repository *Repository) UpdateAchievement(ctx context.Context, value progressionrecord.AchievementDefinition, version int64) (progressionrecord.AchievementDefinition, error) {
	err := repository.executorFor(ctx).QueryRow(ctx, `update achievement_definitions set category=$2,subcategory=$3,trigger_key=$4,visible=$5,enabled=$6,version=version+1,updated_at=now() where id=$1 and version=$7 returning name,version`, value.ID, value.Category, value.Subcategory, value.TriggerKey, value.Visible, value.Enabled, version).Scan(&value.Name, &value.Version)
	if noRows(err) {
		err = repository.missingOrConflict(ctx, "achievement_definitions", "id", value.ID)
	}
	return value, err
}

// DisableAchievement soft-disables one achievement definition.
func (repository *Repository) DisableAchievement(ctx context.Context, id int64) (bool, error) {
	result, err := repository.executorFor(ctx).Exec(ctx, `update achievement_definitions set enabled=false,version=version+1,updated_at=now() where id=$1 and enabled`, id)
	return result.RowsAffected() > 0, err
}

// UpsertAchievementLevel creates or replaces one cumulative level.
func (repository *Repository) UpsertAchievementLevel(ctx context.Context, value progressionrecord.AchievementLevel) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `insert into achievement_levels(definition_id,level,progress_needed,reward_currency_type,reward_amount,score_points) values($1,$2,$3,$4,$5,$6) on conflict(definition_id,level) do update set progress_needed=excluded.progress_needed,reward_currency_type=excluded.reward_currency_type,reward_amount=excluded.reward_amount,score_points=excluded.score_points`, value.DefinitionID, value.Level, value.ProgressNeeded, value.RewardCurrencyType, value.RewardAmount, value.ScorePoints)
	return err
}

// DeleteAchievementLevel removes only the current highest level.
func (repository *Repository) DeleteAchievementLevel(ctx context.Context, definitionID int64, level int32) (bool, error) {
	result, err := repository.executorFor(ctx).Exec(ctx, `delete from achievement_levels where definition_id=$1 and level=$2 and level=(select max(level) from achievement_levels where definition_id=$1)`, definitionID, level)
	if err != nil {
		return false, err
	}
	if result.RowsAffected() > 0 {
		return true, nil
	}
	var exists bool
	err = repository.executorFor(ctx).QueryRow(ctx, `select exists(select 1 from achievement_levels where definition_id=$1 and level=$2)`, definitionID, level).Scan(&exists)
	if err != nil {
		return false, err
	}
	if exists {
		return false, progressionrecord.ErrConflict
	}
	return false, progressionrecord.ErrNotFound
}

// achievement returns one definition for validation helpers.
func (repository *Repository) achievement(ctx context.Context, id int64) (progressionrecord.AchievementDefinition, error) {
	var value progressionrecord.AchievementDefinition
	err := repository.executorFor(ctx).QueryRow(ctx, `select id,name,category,subcategory,trigger_key,visible,enabled,version from achievement_definitions where id=$1`, id).Scan(&value.ID, &value.Name, &value.Category, &value.Subcategory, &value.TriggerKey, &value.Visible, &value.Enabled, &value.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		err = progressionrecord.ErrNotFound
	}
	return value, err
}
