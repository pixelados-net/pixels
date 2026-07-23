// Package talent derives talent track levels from achievement progress.
package talent

import (
	"context"

	"github.com/niflaot/pixels/internal/permission"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	progressionengine "github.com/niflaot/pixels/internal/realm/progression/engine"
	progressionobservability "github.com/niflaot/pixels/internal/realm/progression/observability"
	"github.com/niflaot/pixels/internal/realm/progression/policy"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
)

// BadgeGranter grants durable talent badge rewards.
type BadgeGranter interface {
	// GrantBadge grants one badge idempotently.
	GrantBadge(context.Context, int64, string, string) (bool, error)
}

// PerkManager mutates direct perk-bearing permission nodes.
type PerkManager interface {
	// GrantPlayerNode enables one direct player node.
	GrantPlayerNode(context.Context, int64, permission.Node, bool) error
	// RevokePlayerNode removes one direct player node.
	RevokePlayerNode(context.Context, int64, permission.Node) error
}

// ItemGranter grants talent furniture rewards.
type ItemGranter interface {
	// Grant creates inventory furniture.
	Grant(context.Context, furnitureservice.GrantParams) ([]furnituremodel.Item, error)
}

// Projector publishes committed talent level changes.
type Projector interface {
	// LevelUp publishes one newly paid talent level.
	LevelUp(context.Context, int64, progressionrecord.TalentLevel)
}

// Service owns derived talent evaluation and rewards.
type Service struct {
	// catalog owns immutable track requirements.
	catalog *progressionengine.Catalog
	// store persists player progression.
	store progressionrecord.Store
	// badges owns badge rewards.
	badges BadgeGranter
	// items owns furniture rewards.
	items ItemGranter
	// permissions owns perk rewards.
	permissions PerkManager
	// projector publishes client updates.
	projector Projector
	// metrics stores process-wide progression telemetry.
	metrics *progressionobservability.Metrics
}

// SetMetrics attaches process-wide telemetry before serving talent rewards.
func (service *Service) SetMetrics(metrics *progressionobservability.Metrics) {
	service.metrics = metrics
}

// New creates a talent service.
func New(catalog *progressionengine.Catalog, store progressionrecord.Store, badges BadgeGranter, items ItemGranter, permissions PerkManager, projectors ...Projector) *Service {
	service := &Service{catalog: catalog, store: store, badges: badges, items: items, permissions: permissions}
	if len(projectors) > 0 {
		service.projector = projectors[0]
	}
	return service
}

// Recalculate evaluates only tracks referencing the changed achievement.
func (service *Service) Recalculate(ctx context.Context, playerID int64, definitionID int64) error {
	generation := service.catalog.Current()
	if generation == nil || len(generation.TalentTracksByAchievement[definitionID]) == 0 {
		return nil
	}
	progress, err := service.store.PlayerAchievements(ctx, playerID)
	if err != nil {
		return err
	}
	levels := make(map[int64]int32, len(progress))
	for _, value := range progress {
		levels[value.DefinitionID] = value.Level
	}
	paid, err := service.store.PlayerTalents(ctx, playerID)
	if err != nil {
		return err
	}
	current := make(map[string]int32, len(paid))
	for _, value := range paid {
		current[value.Track] = value.Level
	}
	for _, track := range generation.TalentTracksByAchievement[definitionID] {
		if err = service.advance(ctx, playerID, track, levels, current[track]); err != nil {
			return err
		}
	}
	return nil
}

// Levels returns one track's immutable level definitions.
func (service *Service) Levels(track string) []progressionrecord.TalentLevel {
	generation := service.catalog.Current()
	if generation == nil {
		return nil
	}
	return generation.TalentByTrack[track]
}

// PlayerLevels returns paid talent levels.
func (service *Service) PlayerLevels(ctx context.Context, playerID int64) ([]progressionrecord.PlayerTalent, error) {
	return service.store.PlayerTalents(ctx, playerID)
}

// MeetsGuideLevel reports whether the paid helpers track meets a configured minimum.
func (service *Service) MeetsGuideLevel(ctx context.Context, playerID int64, minimum int32) (bool, error) {
	if minimum <= 0 {
		return true, nil
	}
	levels, err := service.store.PlayerTalents(ctx, playerID)
	if err != nil {
		return false, err
	}
	for _, level := range levels {
		if level.Track == "helpers" {
			return level.Level >= minimum, nil
		}
	}
	return false, nil
}

// advance pays all consecutively satisfied levels.
func (service *Service) advance(ctx context.Context, playerID int64, track string, achievements map[int64]int32, current int32) error {
	for _, level := range service.Levels(track) {
		if level.Level <= current || !satisfied(level, achievements) {
			continue
		}
		if level.Level != current+1 {
			break
		}
		advanced := false
		err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
			var advanceErr error
			advanced, advanceErr = service.store.SetTalent(txCtx, playerID, track, level.Level)
			if advanceErr != nil || !advanced {
				return advanceErr
			}
			return service.reward(txCtx, playerID, level)
		})
		if err != nil {
			return err
		}
		if !advanced {
			current = level.Level
			continue
		}
		current = level.Level
		service.recordRewards(level)
		if service.projector != nil {
			service.projector.LevelUp(ctx, playerID, level)
		}
	}
	return nil
}

// reward grants all rewards from one paid talent level.
func (service *Service) reward(ctx context.Context, playerID int64, level progressionrecord.TalentLevel) error {
	for _, badge := range level.RewardBadges {
		if _, err := service.badges.GrantBadge(ctx, playerID, badge, "talent"); err != nil {
			return err
		}
	}
	for _, item := range level.RewardItems {
		if _, err := service.items.Grant(ctx, furnitureservice.GrantParams{DefinitionID: item, OwnerPlayerID: playerID, Quantity: 1, ExtraData: "0"}); err != nil {
			return err
		}
	}
	for _, perk := range level.RewardPerks {
		if perk == "TRADE" {
			if err := service.permissions.GrantPlayerNode(ctx, playerID, policy.TradePerk, true); err != nil {
				return err
			}
		}
	}
	return nil
}

// recordRewards updates telemetry after one talent-level transaction commits.
func (service *Service) recordRewards(level progressionrecord.TalentLevel) {
	for range level.RewardBadges {
		service.metrics.RecordReward("talent.badge")
	}
	for range level.RewardItems {
		service.metrics.RecordReward("talent.item")
	}
	for _, perk := range level.RewardPerks {
		if perk == "TRADE" {
			service.metrics.RecordReward("talent.perk")
		}
	}
}

// satisfied reports whether every achievement requirement is met.
func satisfied(level progressionrecord.TalentLevel, achievements map[int64]int32) bool {
	for _, requirement := range level.Requirements {
		if achievements[requirement.DefinitionID] < requirement.RequiredLevel {
			return false
		}
	}
	return true
}
