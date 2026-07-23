package progression

import (
	"context"
	"testing"

	progressionengine "github.com/niflaot/pixels/internal/realm/progression/engine"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// emptyCatalogSource returns an unloaded catalog for missing-group behavior.
type emptyCatalogSource struct{}

// Catalog returns no durable definitions.
func (emptyCatalogSource) Catalog(context.Context) (progressionrecord.Catalog, error) {
	return progressionrecord.Catalog{}, nil
}

// TestExecuteProgressionRejectsMalformedConfiguration verifies effects fail closed.
func TestExecuteProgressionRejectsMalformedConfiguration(t *testing.T) {
	service := New(nil, progressionengine.NewCatalog(emptyCatalogSource{}), nil)
	tests := []struct {
		// name identifies the validation case.
		name string
		// operation identifies the effect request.
		operation effect.ProgressionOperation
		// node stores compiled WIRED parameters.
		node *configuration.Node
		// playerID identifies the actor.
		playerID int64
		// want stores the expected result status.
		want effect.Status
	}{
		{"missing actor", effect.ProgressAchievement, &configuration.Node{Parameters: configuration.Parameters{Text: "Badge"}}, 0, effect.Skipped},
		{"missing node", effect.ProgressAchievement, nil, 7, effect.Skipped},
		{"empty achievement", effect.ProgressAchievement, &configuration.Node{}, 7, effect.Blocked},
		{"unknown achievement", effect.ProgressAchievement, &configuration.Node{Parameters: configuration.Parameters{Text: "Missing"}}, 7, effect.Skipped},
		{"invalid quest", effect.ProgressQuest, &configuration.Node{Parameters: configuration.Parameters{Text: "no"}}, 7, effect.Blocked},
		{"invalid operation", effect.ProgressionOperation(99), &configuration.Node{}, 7, effect.Blocked},
	}
	for _, test := range tests {
		result, err := service.ExecuteProgression(context.Background(), test.operation, test.node, trigger.Event{PlayerID: test.playerID})
		if err != nil || result.Status != test.want {
			t.Errorf("%s status=%d want=%d err=%v", test.name, result.Status, test.want, err)
		}
	}
}
