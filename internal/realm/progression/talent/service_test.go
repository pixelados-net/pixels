package talent

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/internal/permission"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	progressionengine "github.com/niflaot/pixels/internal/realm/progression/engine"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
)

// talentCatalogSource returns one fixed talent catalog.
type talentCatalogSource struct{ value progressionrecord.Catalog }

// Catalog returns the fixed talent catalog.
func (source talentCatalogSource) Catalog(context.Context) (progressionrecord.Catalog, error) {
	return source.value, nil
}

// memoryTalentStore implements focused derived-level persistence.
type memoryTalentStore struct {
	progressionrecord.Store
	achievementCalls int
	talentCalls      int
	achievementLevel int32
	level            int32
}

// WithinTransaction executes one in-memory level transaction.
func (*memoryTalentStore) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	return work(ctx)
}

// PlayerAchievements returns one prerequisite level.
func (store *memoryTalentStore) PlayerAchievements(context.Context, int64) ([]progressionrecord.PlayerAchievement, error) {
	store.achievementCalls++
	return []progressionrecord.PlayerAchievement{{DefinitionID: 1, Level: store.achievementLevel}}, nil
}

// PlayerTalents returns the current paid level.
func (store *memoryTalentStore) PlayerTalents(context.Context, int64) ([]progressionrecord.PlayerTalent, error) {
	store.talentCalls++
	if store.level == 0 {
		return nil, nil
	}
	return []progressionrecord.PlayerTalent{{Track: "citizenship", Level: store.level}}, nil
}

// SetTalent advances the current paid level.
func (store *memoryTalentStore) SetTalent(_ context.Context, _ int64, _ string, level int32) (bool, error) {
	if level <= store.level {
		return false, nil
	}
	store.level = level
	return true, nil
}

// ForceTalent replaces the current paid level.
func (store *memoryTalentStore) ForceTalent(_ context.Context, _ int64, _ string, level int32) error {
	store.level = level
	return nil
}

// talentBadges records talent badge grants.
type talentBadges struct{ count int }

// GrantBadge records one badge grant.
func (badges *talentBadges) GrantBadge(context.Context, int64, string, string) (bool, error) {
	badges.count++
	return true, nil
}

// talentItems records talent furniture grants.
type talentItems struct{ count int }

// Grant records one furniture grant.
func (items *talentItems) Grant(context.Context, furnitureservice.GrantParams) ([]furnituremodel.Item, error) {
	items.count++
	return []furnituremodel.Item{{}}, nil
}

// talentPerks records direct permission mutations.
type talentPerks struct {
	granted int
	revoked int
}

// GrantPlayerNode records one perk grant.
func (perks *talentPerks) GrantPlayerNode(context.Context, int64, permission.Node, bool) error {
	perks.granted++
	return nil
}

// RevokePlayerNode records one perk removal.
func (perks *talentPerks) RevokePlayerNode(context.Context, int64, permission.Node) error {
	perks.revoked++
	return nil
}

// talentProjector records paid level projections.
type talentProjector struct{ levels []int32 }

// LevelUp records one projected level.
func (projector *talentProjector) LevelUp(_ context.Context, _ int64, level progressionrecord.TalentLevel) {
	projector.levels = append(projector.levels, level.Level)
}

// talentFixture builds one loaded citizenship track.
func talentFixture(t testing.TB) (*Service, *memoryTalentStore, *talentBadges, *talentItems, *talentPerks, *talentProjector) {
	t.Helper()
	levels := []progressionrecord.TalentLevel{
		{Track: "citizenship", Level: 1, Requirements: []progressionrecord.TalentRequirement{{DefinitionID: 1, RequiredLevel: 1}}, RewardBadges: []string{"ONE"}},
		{Track: "citizenship", Level: 2, Requirements: []progressionrecord.TalentRequirement{{DefinitionID: 1, RequiredLevel: 2}}, RewardItems: []int64{10}, RewardPerks: []string{"TRADE"}},
	}
	catalog := progressionengine.NewCatalog(talentCatalogSource{value: progressionrecord.Catalog{Talents: levels}})
	if err := catalog.Reload(context.Background()); err != nil {
		t.Fatal(err)
	}
	store := &memoryTalentStore{achievementLevel: 2}
	badges, items, perks, projector := &talentBadges{}, &talentItems{}, &talentPerks{}, &talentProjector{}
	return New(catalog, store, badges, items, perks, projector), store, badges, items, perks, projector
}

// TestIrrelevantAchievementDoesNoWork verifies the reverse index protects the hot path.
func TestIrrelevantAchievementDoesNoWork(t *testing.T) {
	service, store, _, _, _, _ := talentFixture(t)
	if err := service.Recalculate(context.Background(), 42, 999); err != nil {
		t.Fatal(err)
	}
	if store.achievementCalls != 0 || store.talentCalls != 0 {
		t.Fatalf("achievement queries=%d talent queries=%d", store.achievementCalls, store.talentCalls)
	}
}

// TestRecalculateCrossesConsecutiveLevels verifies all newly satisfied rewards are paid.
func TestRecalculateCrossesConsecutiveLevels(t *testing.T) {
	service, store, badges, items, perks, projector := talentFixture(t)
	if err := service.Recalculate(context.Background(), 42, 1); err != nil {
		t.Fatal(err)
	}
	if store.level != 2 || badges.count != 1 || items.count != 1 || perks.granted != 1 {
		t.Fatalf("level=%d badges=%d items=%d perks=%d", store.level, badges.count, items.count, perks.granted)
	}
	if len(projector.levels) != 2 || projector.levels[0] != 1 || projector.levels[1] != 2 {
		t.Fatalf("projections %v", projector.levels)
	}
}

// TestForceDownRevokesTradePerk verifies corrective track rollback aligns the perk.
func TestForceDownRevokesTradePerk(t *testing.T) {
	service, store, _, _, perks, _ := talentFixture(t)
	store.level = 2
	if err := service.Force(context.Background(), 42, "citizenship", 1); err != nil {
		t.Fatal(err)
	}
	if store.level != 1 || perks.revoked != 1 {
		t.Fatalf("level=%d revoked=%d", store.level, perks.revoked)
	}
}
