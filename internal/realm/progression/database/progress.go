package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
)

// PlayerAchievements lists one player's durable progress.
func (repository *Repository) PlayerAchievements(ctx context.Context, playerID int64) ([]progressionrecord.PlayerAchievement, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select player_id,definition_id,progress,level,last_daily_at from player_achievement_progress where player_id=$1 order by definition_id`, playerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := make([]progressionrecord.PlayerAchievement, 0)
	for rows.Next() {
		value, scanErr := scanProgress(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		values = append(values, value)
	}
	return values, rows.Err()
}

// MutateProgress locks and increments one progress row atomically.
func (repository *Repository) MutateProgress(ctx context.Context, playerID int64, definition progressionrecord.AchievementDefinition, amount int64, daily bool) (progressionrecord.ProgressMutation, error) {
	executor := repository.executorFor(ctx)
	if _, err := executor.Exec(ctx, `insert into player_achievement_progress(player_id,definition_id) values($1,$2) on conflict do nothing`, playerID, definition.ID); err != nil {
		return progressionrecord.ProgressMutation{}, err
	}
	before, err := scanProgress(executor.QueryRow(ctx, `select player_id,definition_id,progress,level,last_daily_at from player_achievement_progress where player_id=$1 and definition_id=$2 for update`, playerID, definition.ID))
	if err != nil {
		return progressionrecord.ProgressMutation{}, err
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	if daily && before.LastDailyAt != nil && before.LastDailyAt.Equal(today) {
		return progressionrecord.ProgressMutation{Before: before, After: before}, nil
	}
	after := before
	after.Progress += amount
	if count := len(definition.Levels); count > 0 && after.Progress > definition.Levels[count-1].ProgressNeeded {
		after.Progress = definition.Levels[count-1].ProgressNeeded
	}
	after.Level, _ = levelFor(definition.Levels, after.Progress)
	if daily {
		after.LastDailyAt = &today
	}
	_, err = executor.Exec(ctx, `update player_achievement_progress set progress=$3,level=$4,last_daily_at=$5,updated_at=now() where player_id=$1 and definition_id=$2`, playerID, definition.ID, after.Progress, after.Level, after.LastDailyAt)
	return progressionrecord.ProgressMutation{Before: before, After: after, Crossed: crossedLevels(definition.Levels, before.Level, after.Level)}, err
}

// SetProgress forces one progress row and returns newly crossed levels.
func (repository *Repository) SetProgress(ctx context.Context, playerID int64, definition progressionrecord.AchievementDefinition, progress int64, level int32, payRewards bool) (progressionrecord.ProgressMutation, error) {
	executor := repository.executorFor(ctx)
	if _, err := executor.Exec(ctx, `insert into player_achievement_progress(player_id,definition_id) values($1,$2) on conflict do nothing`, playerID, definition.ID); err != nil {
		return progressionrecord.ProgressMutation{}, err
	}
	before, err := scanProgress(executor.QueryRow(ctx, `select player_id,definition_id,progress,level,last_daily_at from player_achievement_progress where player_id=$1 and definition_id=$2 for update`, playerID, definition.ID))
	if err != nil {
		return progressionrecord.ProgressMutation{}, err
	}
	after := before
	after.Progress, after.Level = progress, level
	if _, err = executor.Exec(ctx, `update player_achievement_progress set progress=$3,level=$4,updated_at=now() where player_id=$1 and definition_id=$2`, playerID, definition.ID, progress, level); err != nil {
		return progressionrecord.ProgressMutation{}, err
	}
	crossed := []progressionrecord.AchievementLevel(nil)
	if payRewards {
		crossed = crossedLevels(definition.Levels, before.Level, after.Level)
	}
	return progressionrecord.ProgressMutation{Before: before, After: after, Crossed: crossed}, nil
}

// ResetProgress deletes one player's progress row.
func (repository *Repository) ResetProgress(ctx context.Context, playerID int64, definitionID int64) (bool, error) {
	result, err := repository.executorFor(ctx).Exec(ctx, `delete from player_achievement_progress where player_id=$1 and definition_id=$2`, playerID, definitionID)
	return result.RowsAffected() > 0, err
}

// AddScore increments one player's durable achievement score.
func (repository *Repository) AddScore(ctx context.Context, playerID int64, amount int32) (int32, error) {
	var score int32
	err := repository.executorFor(ctx).QueryRow(ctx, `update players set achievement_score=achievement_score+$2,updated_at=now(),version=version+1 where id=$1 returning achievement_score`, playerID, amount).Scan(&score)
	return score, err
}

// scanProgress scans one progress row.
func scanProgress(row interface{ Scan(...any) error }) (progressionrecord.PlayerAchievement, error) {
	var value progressionrecord.PlayerAchievement
	var daily pgtype.Date
	err := row.Scan(&value.PlayerID, &value.DefinitionID, &value.Progress, &value.Level, &daily)
	if daily.Valid {
		value.LastDailyAt = &daily.Time
	}
	return value, err
}

// levelFor resolves the highest crossed cumulative threshold.
func levelFor(levels []progressionrecord.AchievementLevel, progress int64) (int32, int64) {
	var level int32
	var next int64
	for _, candidate := range levels {
		if progress < candidate.ProgressNeeded {
			next = candidate.ProgressNeeded
			break
		}
		level = candidate.Level
	}
	return level, next
}

// crossedLevels returns levels paid by one monotonic transition.
func crossedLevels(levels []progressionrecord.AchievementLevel, before int32, after int32) []progressionrecord.AchievementLevel {
	values := make([]progressionrecord.AchievementLevel, 0, after-before)
	for _, level := range levels {
		if level.Level > before && level.Level <= after {
			values = append(values, level)
		}
	}
	return values
}
