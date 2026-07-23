package engine

import (
	"context"
	"errors"
	"fmt"

	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
	"github.com/niflaot/pixels/pkg/postgres"
)

// validateDependencies reports incomplete engine wiring.
func (service *Service) validateDependencies() error {
	if service.catalog == nil || service.store == nil || service.badges == nil {
		return errors.New("progression engine dependencies unavailable")
	}
	return nil
}

// apply executes one achievement transition atomically.
func (service *Service) apply(ctx context.Context, playerID int64, definition progressionrecord.AchievementDefinition, amount int64, daily bool) error {
	var transition Transition
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		mutation, err := service.store.MutateProgress(txCtx, playerID, definition, amount, daily)
		if err != nil {
			return err
		}
		transition = Transition{PlayerID: playerID, Definition: definition, Mutation: mutation}
		for _, level := range mutation.Crossed {
			if err = service.rewardLevel(txCtx, playerID, definition, level); err != nil {
				return err
			}
			if level.ScorePoints > 0 {
				transition.Score, err = service.store.AddScore(txCtx, playerID, level.ScorePoints)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	service.project(ctx, transition)
	service.observeProgress(playerID, definition.ID, transition.Mutation.After.Progress)
	service.recordAchievementRewards(transition.Mutation.Crossed)
	if len(transition.Mutation.Crossed) > 0 && service.talents != nil {
		return service.talents.Recalculate(ctx, playerID, definition.ID)
	}
	return nil
}

// project publishes one transition after its outer transaction commits.
func (service *Service) project(ctx context.Context, transition Transition) {
	if service.projector == nil {
		return
	}
	if !postgres.AfterCommit(ctx, func(committed context.Context) { service.projector.Project(committed, transition) }) {
		service.projector.Project(ctx, transition)
	}
}

// rewardLevel grants one badge transition and optional currency reward.
func (service *Service) rewardLevel(ctx context.Context, playerID int64, definition progressionrecord.AchievementDefinition, level progressionrecord.AchievementLevel) error {
	newCode := fmt.Sprintf("ACH_%s%d", definition.Name, level.Level)
	if level.Level == 1 {
		if _, err := service.badges.GrantBadge(ctx, playerID, newCode, "achievement"); err != nil {
			return err
		}
	} else if err := service.replaceBadge(ctx, playerID, definition.Name, level.Level, newCode); err != nil {
		return err
	}
	if level.RewardAmount <= 0 || service.currencies == nil {
		return nil
	}
	_, err := service.currencies.Grant(ctx, currencyservice.GrantParams{PlayerID: playerID, CurrencyType: level.RewardCurrencyType, Amount: level.RewardAmount, Reason: "achievement level reward", ActorKind: currencyservice.ActorSystem})
	return err
}

// recordAchievementRewards updates telemetry only after the mutation commits.
func (service *Service) recordAchievementRewards(levels []progressionrecord.AchievementLevel) {
	service.metrics.RecordLevelUps(len(levels))
	for _, level := range levels {
		service.metrics.RecordReward("achievement.badge")
		if level.RewardAmount > 0 && service.currencies != nil {
			service.metrics.RecordReward("achievement.currency")
		}
	}
}
