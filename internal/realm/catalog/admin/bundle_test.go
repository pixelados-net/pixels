package admin

import (
	"context"
	"testing"

	roombundle "github.com/niflaot/pixels/internal/realm/room/record/bundle"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
)

// adminRoomBundles supplies marked template validation.
type adminRoomBundles struct{}

// Clone is unused by catalog administration.
func (adminRoomBundles) Clone(context.Context, roombundle.CloneParams) (roombundle.CloneResult, error) {
	return roombundle.CloneResult{}, nil
}

// Preview returns one template product.
func (adminRoomBundles) Preview(context.Context, int64) ([]roombundle.Product, error) {
	return []roombundle.Product{{DefinitionID: 2, Quantity: 2}}, nil
}

// Mark is unused by catalog administration.
func (adminRoomBundles) Mark(context.Context, int64) (roommodel.Room, error) {
	return roommodel.Room{}, nil
}

// Unmark is unused by catalog administration.
func (adminRoomBundles) Unmark(context.Context, int64) (roommodel.Room, error) {
	return roommodel.Room{}, nil
}

// Templates is unused by catalog administration.
func (adminRoomBundles) Templates(context.Context) ([]roommodel.Room, error) { return nil, nil }

// FindTemplate returns a marked room template.
func (adminRoomBundles) FindTemplate(context.Context, int64) (roommodel.Room, bool, error) {
	return roommodel.Room{IsBundleTemplate: true}, true, nil
}

// TestCreateRoomBundleItemAcceptsTemplateWithoutDefinition verifies offer exclusivity.
func TestCreateRoomBundleItemAcceptsTemplateWithoutDefinition(t *testing.T) {
	fixture := newFixture()
	templateID := int64(100)
	fixture.service.WithRoomBundles(adminRoomBundles{})
	item, err := fixture.service.CreateItem(context.Background(), ItemInput{PageID: 1, RoomBundleTemplateRoomID: &templateID, Name: "starter_loft_bundle", CostCredits: 75, PointsType: -1, Enabled: true})
	if err != nil || !item.IsRoomBundle() || item.DefinitionID != 0 || item.Amount != 0 {
		t.Fatalf("item=%#v error=%v", item, err)
	}
}
