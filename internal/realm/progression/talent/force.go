package talent

import (
	"context"

	"github.com/niflaot/pixels/internal/realm/progression/policy"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
)

// Force replaces one player's exact track level and pays newly crossed rewards.
func (service *Service) Force(ctx context.Context, playerID int64, track string, target int32) error {
	levels := service.Levels(track)
	if target < 0 || target > 0 && (len(levels) == 0 || levels[len(levels)-1].Level < target) {
		return progressionrecord.ErrInvalid
	}
	current, err := service.currentLevel(ctx, playerID, track)
	if err != nil {
		return err
	}
	err = service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		if forceErr := service.store.ForceTalent(txCtx, playerID, track, target); forceErr != nil {
			return forceErr
		}
		for _, level := range levels {
			if level.Level > current && level.Level <= target {
				if forceErr := service.reward(txCtx, playerID, level); forceErr != nil {
					return forceErr
				}
			}
		}
		return service.syncPerks(txCtx, playerID, levels, target)
	})
	if err == nil && target > current {
		for _, level := range levels {
			if level.Level > current && level.Level <= target {
				service.recordRewards(level)
				if service.projector != nil {
					service.projector.LevelUp(ctx, playerID, level)
				}
			}
		}
	}
	return err
}

// currentLevel returns one player's paid track level.
func (service *Service) currentLevel(ctx context.Context, playerID int64, track string) (int32, error) {
	values, err := service.store.PlayerTalents(ctx, playerID)
	if err != nil {
		return 0, err
	}
	for _, value := range values {
		if value.Track == track {
			return value.Level, nil
		}
	}
	return 0, nil
}

// syncPerks aligns supported direct perks with one forced track level.
func (service *Service) syncPerks(ctx context.Context, playerID int64, levels []progressionrecord.TalentLevel, target int32) error {
	trade := false
	for _, level := range levels {
		if level.Level > target {
			break
		}
		for _, perk := range level.RewardPerks {
			trade = trade || perk == "TRADE"
		}
	}
	if trade {
		return service.permissions.GrantPlayerNode(ctx, playerID, policy.TradePerk, true)
	}
	return service.permissions.RevokePlayerNode(ctx, playerID, policy.TradePerk)
}
