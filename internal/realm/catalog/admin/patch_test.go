package admin

import (
	"context"
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/permission"
	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestApplyPagePatchAppliesEveryPresentField verifies complete page patch mapping.
func TestApplyPagePatchAppliesEveryPresentField(t *testing.T) {
	parentID := int64(2)
	name := "chairs"
	layout := "spaces"
	color := int32(3)
	image := int32(4)
	node := permission.Node("catalog.admin.manage")
	order := int32(5)
	visible := true
	enabled := true
	club := true
	parent := &parentID
	required := &node
	page := catalogmodel.Page{}
	applyPagePatch(&page, PagePatch{ParentID: &parent, Name: &name, Layout: &layout, IconColor: &color,
		IconImage: &image, RequiredNode: &required, OrderNum: &order, Visible: &visible, Enabled: &enabled, ClubOnly: &club})
	if page.ParentID == nil || *page.ParentID != 2 || page.Name != name || page.Layout != layout || page.IconColor != color ||
		page.IconImage != image || page.RequiredNode == nil || *page.RequiredNode != node || page.OrderNum != order || !page.Visible || !page.Enabled || !page.ClubOnly {
		t.Fatalf("unexpected page %#v", page)
	}
}

// TestApplyItemPatchAppliesEveryPresentField verifies complete offer patch mapping.
func TestApplyItemPatchAppliesEveryPresentField(t *testing.T) {
	pageID := int64(2)
	definitionID := int64(3)
	name := "chair"
	credits := int64(4)
	points := int64(0)
	pointsType := int32(-1)
	amount := int32(2)
	stack := int32(10)
	bundle := true
	giftable := true
	club := true
	order := int32(8)
	enabled := true
	extra := "state"
	item := catalogmodel.Item{}
	applyItemPatch(&item, ItemPatch{PageID: &pageID, DefinitionID: &definitionID, Name: &name, CostCredits: &credits,
		CostPoints: &points, PointsType: &pointsType, Amount: &amount, LimitedStack: &stack,
		BundleDiscountEnabled: &bundle, Giftable: &giftable,
		ClubOnly: &club, OrderNum: &order, Enabled: &enabled, ExtraData: &extra})
	if item.PageID != pageID || item.DefinitionID != definitionID || item.Name != name || item.CostCredits != credits ||
		item.PointsType != pointsType || item.Amount != amount || item.LimitedStack != stack || !item.BundleDiscountEnabled ||
		!item.Giftable || !item.ClubOnly || item.OrderNum != order || !item.Enabled || item.ExtraData != extra {
		t.Fatalf("unexpected item %#v", item)
	}
}

// TestAdministrationReadCapabilities verifies refresh and sanitize delegation.
func TestAdministrationReadCapabilities(t *testing.T) {
	fixture := newFixture()
	if err := fixture.service.Refresh(context.Background()); err != nil || fixture.catalog.refreshes != 1 {
		t.Fatalf("unexpected refreshes=%d error %v", fixture.catalog.refreshes, err)
	}
	definitions, err := fixture.service.SanitizeList(context.Background())
	if err != nil || len(definitions) != 1 || definitions[0].ID != 9 {
		t.Fatalf("unexpected definitions %#v error %v", definitions, err)
	}
}

// TestAdministrationRejectsInvalidMutations verifies basic mutation guards.
func TestAdministrationRejectsInvalidMutations(t *testing.T) {
	fixture := newFixture()
	if _, err := fixture.service.UpdatePage(context.Background(), 0, PagePatch{}); !errors.Is(err, ErrInvalidPage) {
		t.Fatalf("expected invalid page, got %v", err)
	}
	if _, err := fixture.service.UpdateItem(context.Background(), 0, ItemPatch{}); !errors.Is(err, ErrInvalidItem) {
		t.Fatalf("expected invalid item, got %v", err)
	}
	if err := fixture.service.DeleteItem(context.Background(), 99); !errors.Is(err, ErrItemNotFound) {
		t.Fatalf("expected missing item, got %v", err)
	}
	if _, err := fixture.service.UpdatePage(context.Background(), 99, PagePatch{}); !errors.Is(err, ErrPageNotFound) {
		t.Fatalf("expected missing page, got %v", err)
	}
	fixture.store.items = []catalogmodel.Item{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, PageID: 1,
		DefinitionID: 2, Name: "ltd", PointsType: -1, Amount: 1, LimitedStack: 2, LimitedSells: 2}}
	stack := int32(1)
	if _, err := fixture.service.UpdateItem(context.Background(), 1, ItemPatch{LimitedStack: &stack}); !errors.Is(err, ErrLimitedBelowSales) {
		t.Fatalf("expected limited stack error, got %v", err)
	}
	_, err := fixture.service.CreateItem(context.Background(), ItemInput{PageID: 1, DefinitionID: 2, Name: "mixed",
		CostCredits: 1, CostPoints: 1, PointsType: 5, Amount: 1})
	if !errors.Is(err, ErrInvalidItem) {
		t.Fatalf("expected invalid mixed price, got %v", err)
	}
}
