// Package progression adapts WIRED effects to durable progression services.
package progression

import (
	"context"
	"strconv"
	"strings"

	progressionengine "github.com/niflaot/pixels/internal/realm/progression/engine"
	progressionquest "github.com/niflaot/pixels/internal/realm/progression/quest"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// Service executes durable progression WIRED effects.
type Service struct {
	// achievements progresses named achievement groups.
	achievements *progressionengine.Service
	// catalog resolves immutable achievement group names.
	catalog *progressionengine.Catalog
	// quests owns quest activation and progress.
	quests *progressionquest.Service
}

// New creates a progression WIRED effect adapter.
func New(achievements *progressionengine.Service, catalog *progressionengine.Catalog, quests *progressionquest.Service) *Service {
	return &Service{achievements: achievements, catalog: catalog, quests: quests}
}

// ExecuteProgression executes one configured player mutation.
func (service *Service) ExecuteProgression(ctx context.Context, operation effect.ProgressionOperation, node *configuration.Node, event trigger.Event) (effect.Result, error) {
	if event.PlayerID <= 0 || node == nil {
		return effect.Result{Status: effect.Skipped}, nil
	}
	text := strings.TrimSpace(node.Parameters.Text)
	switch operation {
	case effect.ProgressAchievement:
		if text == "" {
			return effect.Result{Status: effect.Blocked}, nil
		}
		changed, err := service.progressAchievement(ctx, event.PlayerID, text)
		if err != nil {
			return effect.Result{Status: effect.Blocked}, err
		}
		if !changed {
			return effect.Result{Status: effect.Skipped}, nil
		}
	case effect.ProgressQuest, effect.StartQuest:
		questID, err := strconv.ParseInt(text, 10, 64)
		if err != nil || questID <= 0 {
			return effect.Result{Status: effect.Blocked}, nil
		}
		if operation == effect.StartQuest {
			err = service.quests.Activate(ctx, event.PlayerID, questID)
		} else {
			var changed bool
			changed, err = service.progressQuest(ctx, event.PlayerID, questID)
			if err == nil && !changed {
				return effect.Result{Status: effect.Skipped}, nil
			}
		}
		if err != nil {
			return effect.Result{Status: effect.Blocked}, err
		}
	default:
		return effect.Result{Status: effect.Blocked}, nil
	}
	return effect.Result{Status: effect.Applied}, nil
}

// progressAchievement advances one exact cached group name.
func (service *Service) progressAchievement(ctx context.Context, playerID int64, name string) (bool, error) {
	generation := service.catalog.Current()
	if generation == nil {
		return false, nil
	}
	for _, definition := range generation.Catalog.Achievements {
		if definition.Enabled && strings.EqualFold(definition.Name, name) {
			return true, service.achievements.ProgressDefinition(ctx, playerID, definition.ID, 1)
		}
	}
	return false, nil
}

// progressQuest advances only the configured active quest.
func (service *Service) progressQuest(ctx context.Context, playerID int64, questID int64) (bool, error) {
	quest, _, found, err := service.quests.Active(ctx, playerID)
	if err != nil || !found || quest.ID != questID {
		return false, err
	}
	return true, service.quests.ProgressTrigger(ctx, playerID, quest.TriggerKey, quest.GoalData, 1)
}

var serviceAssertion effect.ProgressionService = (*Service)(nil)
