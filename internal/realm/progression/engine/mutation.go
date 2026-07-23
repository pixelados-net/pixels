package engine

import (
	"context"
	"fmt"

	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
)

// SetQuestProgressor links the quest engine after dependency construction.
func (service *Service) SetQuestProgressor(progressor QuestProgressor) { service.quests = progressor }

// SetTalentRecalculator links the talent engine after dependency construction.
func (service *Service) SetTalentRecalculator(recalculator TalentRecalculator) {
	service.talents = recalculator
}

// ProgressDefinition applies an administrative delta through the real reward engine.
func (service *Service) ProgressDefinition(ctx context.Context, playerID int64, definitionID int64, amount int64) error {
	definition, err := service.definition(definitionID)
	if err != nil || amount <= 0 {
		return err
	}
	return service.apply(ctx, playerID, definition, amount, false)
}

// SetLevel forces one exact achievement level with optional crossed rewards.
func (service *Service) SetLevel(ctx context.Context, playerID int64, definitionID int64, level int32, payRewards bool) error {
	definition, err := service.definition(definitionID)
	if err != nil {
		return err
	}
	progress, valid := progressForLevel(definition.Levels, level)
	if !valid {
		return progressionrecord.ErrInvalid
	}
	return service.set(ctx, playerID, definition, progress, level, payRewards)
}

// SetTriggerProgress raises matching achievements to an absolute progress value.
func (service *Service) SetTriggerProgress(ctx context.Context, playerID int64, key string, progress int64) error {
	if service == nil || !service.config.Enabled || progress < 0 {
		return nil
	}
	for _, definition := range service.catalog.Achievements(key) {
		level := levelForProgress(definition.Levels, progress)
		if err := service.set(ctx, playerID, *definition, progress, level, true); err != nil {
			return err
		}
	}
	return nil
}

// ResetDefinition deletes one progress row without implicitly removing badges.
func (service *Service) ResetDefinition(ctx context.Context, playerID int64, definitionID int64) (bool, error) {
	if _, err := service.definition(definitionID); err != nil {
		return false, err
	}
	changed, err := service.store.ResetProgress(ctx, playerID, definitionID)
	if err == nil && changed {
		service.observeProgress(playerID, definitionID, 0)
	}
	return changed, err
}

// definition resolves one cached achievement definition.
func (service *Service) definition(definitionID int64) (progressionrecord.AchievementDefinition, error) {
	generation := service.catalog.Current()
	if generation == nil || generation.AchievementByID[definitionID] == nil {
		return progressionrecord.AchievementDefinition{}, progressionrecord.ErrNotFound
	}
	return *generation.AchievementByID[definitionID], nil
}

// set applies one exact progress state atomically.
func (service *Service) set(ctx context.Context, playerID int64, definition progressionrecord.AchievementDefinition, progress int64, level int32, payRewards bool) error {
	var transition Transition
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		mutation, setErr := service.store.SetProgress(txCtx, playerID, definition, progress, level, payRewards)
		if setErr != nil {
			return setErr
		}
		transition = Transition{PlayerID: playerID, Definition: definition, Mutation: mutation}
		for _, crossed := range mutation.Crossed {
			if setErr = service.rewardLevel(txCtx, playerID, definition, crossed); setErr != nil {
				return setErr
			}
			if crossed.ScorePoints > 0 {
				transition.Score, setErr = service.store.AddScore(txCtx, playerID, crossed.ScorePoints)
				if setErr != nil {
					return setErr
				}
			}
		}
		if len(mutation.Crossed) == 0 && mutation.Before.Level != mutation.After.Level {
			return service.syncBadge(txCtx, playerID, definition.Name, mutation.Before.Level, mutation.After.Level)
		}
		return nil
	})
	if err != nil {
		return err
	}
	service.project(ctx, transition)
	service.observeProgress(playerID, definition.ID, transition.Mutation.After.Progress)
	service.recordAchievementRewards(transition.Mutation.Crossed)
	if transition.Mutation.Before.Level != transition.Mutation.After.Level && service.talents != nil {
		return service.talents.Recalculate(ctx, playerID, definition.ID)
	}
	return nil
}

// syncBadge aligns a forced non-reward level transition.
func (service *Service) syncBadge(ctx context.Context, playerID int64, name string, before int32, after int32) error {
	oldCode := ""
	if before > 0 {
		oldCode = fmt.Sprintf("ACH_%s%d", name, before)
	}
	if after == 0 {
		if oldCode == "" {
			return nil
		}
		_, err := service.badges.RemoveBadge(ctx, playerID, oldCode)
		return err
	}
	newCode := fmt.Sprintf("ACH_%s%d", name, after)
	if oldCode != "" {
		replaced, err := service.badges.ReplaceBadge(ctx, playerID, oldCode, newCode, "achievement")
		if err != nil || replaced {
			return err
		}
	}
	_, err := service.badges.GrantBadge(ctx, playerID, newCode, "achievement")
	return err
}

// progressForLevel resolves the exact cumulative threshold for one level.
func progressForLevel(levels []progressionrecord.AchievementLevel, level int32) (int64, bool) {
	if level == 0 {
		return 0, true
	}
	for _, candidate := range levels {
		if candidate.Level == level {
			return candidate.ProgressNeeded, true
		}
	}
	return 0, false
}

// levelForProgress resolves the highest crossed level.
func levelForProgress(levels []progressionrecord.AchievementLevel, progress int64) int32 {
	var level int32
	for _, candidate := range levels {
		if progress < candidate.ProgressNeeded {
			break
		}
		level = candidate.Level
	}
	return level
}

// replaceBadge preserves the prior equipped slot while advancing its code.
func (service *Service) replaceBadge(ctx context.Context, playerID int64, name string, level int32, newCode string) error {
	oldCode := fmt.Sprintf("ACH_%s%d", name, level-1)
	replaced, err := service.badges.ReplaceBadge(ctx, playerID, oldCode, newCode, "achievement")
	if err != nil || replaced {
		return err
	}
	_, err = service.badges.GrantBadge(ctx, playerID, newCode, "achievement")
	return err
}
