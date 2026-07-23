package database

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgtype"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
)

// Catalog loads one complete progression catalog generation.
func (repository *Repository) Catalog(ctx context.Context) (progressionrecord.Catalog, error) {
	catalog := progressionrecord.Catalog{}
	var err error
	if catalog.Achievements, err = repository.achievements(ctx); err != nil {
		return catalog, err
	}
	if catalog.Talents, err = repository.talents(ctx); err != nil {
		return catalog, err
	}
	if catalog.Campaigns, catalog.Quests, err = repository.quests(ctx); err != nil {
		return catalog, err
	}
	if catalog.Quizzes, err = repository.quizzes(ctx); err != nil {
		return catalog, err
	}
	catalog.Promos, err = repository.promos(ctx)
	return catalog, err
}

// achievements loads definitions and levels in one ordered query.
func (repository *Repository) achievements(ctx context.Context) ([]progressionrecord.AchievementDefinition, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select d.id,d.name,d.category,d.subcategory,d.trigger_key,d.visible,d.enabled,d.version,l.level,l.progress_needed,l.reward_currency_type,l.reward_amount,l.score_points from achievement_definitions d left join achievement_levels l on l.definition_id=d.id order by d.id,l.level`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	definitions := make([]progressionrecord.AchievementDefinition, 0)
	var current *progressionrecord.AchievementDefinition
	for rows.Next() {
		var definition progressionrecord.AchievementDefinition
		var level pgtype.Int4
		var needed, reward pgtype.Int8
		var currency, score pgtype.Int4
		if err = rows.Scan(&definition.ID, &definition.Name, &definition.Category, &definition.Subcategory, &definition.TriggerKey, &definition.Visible, &definition.Enabled, &definition.Version, &level, &needed, &currency, &reward, &score); err != nil {
			return nil, err
		}
		if current == nil || current.ID != definition.ID {
			definitions = append(definitions, definition)
			current = &definitions[len(definitions)-1]
		}
		if level.Valid {
			current.Levels = append(current.Levels, progressionrecord.AchievementLevel{DefinitionID: definition.ID, Level: level.Int32, ProgressNeeded: needed.Int64, RewardCurrencyType: currency.Int32, RewardAmount: reward.Int64, ScorePoints: score.Int32})
		}
	}
	return definitions, rows.Err()
}

// talents loads data-driven talent track levels.
func (repository *Repository) talents(ctx context.Context) ([]progressionrecord.TalentLevel, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select track,level,requirements,reward_items,reward_perks,reward_badges from talent_track_levels order by track,level`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	levels := make([]progressionrecord.TalentLevel, 0)
	for rows.Next() {
		var level progressionrecord.TalentLevel
		var requirements []byte
		if err = rows.Scan(&level.Track, &level.Level, &requirements, &level.RewardItems, &level.RewardPerks, &level.RewardBadges); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(requirements, &level.Requirements); err != nil {
			return nil, err
		}
		levels = append(levels, level)
	}
	return levels, rows.Err()
}
