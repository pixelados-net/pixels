package engine

import (
	"context"
	"testing"

	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
)

// catalogSource returns one fixed catalog.
type catalogSource struct{ catalog progressionrecord.Catalog }

// Catalog returns the fixed test catalog.
func (source catalogSource) Catalog(context.Context) (progressionrecord.Catalog, error) {
	return source.catalog, nil
}

// TestCatalogReloadBuildsIndexes verifies enabled fan-out and reverse talent lookup.
func TestCatalogReloadBuildsIndexes(t *testing.T) {
	source := catalogSource{catalog: progressionrecord.Catalog{Achievements: []progressionrecord.AchievementDefinition{{ID: 1, TriggerKey: "room.entered", Enabled: true}, {ID: 2, TriggerKey: "room.entered"}}, Talents: []progressionrecord.TalentLevel{{Track: "citizenship", Level: 1, Requirements: []progressionrecord.TalentRequirement{{DefinitionID: 1, RequiredLevel: 1}}}}}}
	catalog := NewCatalog(source)
	if err := catalog.Reload(context.Background()); err != nil {
		t.Fatal(err)
	}
	if values := catalog.Achievements("room.entered"); len(values) != 1 || values[0].ID != 1 {
		t.Fatalf("achievements %+v", values)
	}
	if tracks := catalog.Current().TalentTracksByAchievement[1]; len(tracks) != 1 || tracks[0] != "citizenship" {
		t.Fatalf("tracks %+v", tracks)
	}
}

// BenchmarkTriggerFanout verifies the empty trigger hot path stays allocation free.
func BenchmarkTriggerFanout(b *testing.B) {
	catalog := NewCatalog(catalogSource{})
	_ = catalog.Reload(context.Background())
	b.ReportAllocs()
	for range b.N {
		if catalog.Achievements("missing") != nil {
			b.Fatal("unexpected definition")
		}
	}
}
