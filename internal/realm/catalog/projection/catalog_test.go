package projection

import (
	"errors"
	"math"
	"testing"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	"github.com/niflaot/pixels/pkg/i18n"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestPageTreeBuildsNestedLocalizedNodes verifies recursive page mapping.
func TestPageTreeBuildsNestedLocalizedNodes(t *testing.T) {
	parentID := int64(1)
	translations := i18n.NewCatalog(i18n.Config{}, map[i18n.Locale]map[i18n.Key]string{"en": {"catalog.page.root": "Furniture"}})
	nodes, err := PageTree([]catalogmodel.Page{
		{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, Name: "root", Visible: true},
		{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 2}}, ParentID: &parentID, Name: "chairs", Visible: true},
	}, translations)
	if err != nil || len(nodes) != 1 || nodes[0].Localization != "Furniture" || len(nodes[0].Children) != 1 {
		t.Fatalf("unexpected nodes %#v error %v", nodes, err)
	}
}

// TestOfferMapsLimitedFloorProduct verifies catalog offer mapping.
func TestOfferMapsLimitedFloorProduct(t *testing.T) {
	item := catalogmodel.Item{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 4}}, Name: "chair", CostCredits: 2,
		Amount: 1, LimitedStack: 10, LimitedSells: 3}
	definition := furnituremodel.Definition{SpriteID: 7, Kind: furnituremodel.KindFloor}
	result, err := Offer(item, definition)
	if err != nil || result.ID != 4 || result.Products[0].LimitedRemaining != 7 || result.Products[0].ClassID != 7 {
		t.Fatalf("unexpected offer %#v error %v", result, err)
	}
}

// TestOfferMapsWallProduct verifies wall definitions use Nitro's wall product discriminator.
func TestOfferMapsWallProduct(t *testing.T) {
	item := catalogmodel.Item{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 8}}, Name: "postit", Amount: 1}
	result, err := Offer(item, furnituremodel.Definition{SpriteID: 1, Kind: furnituremodel.KindWall})
	if err != nil || len(result.Products) != 1 || result.Products[0].Type != "i" {
		t.Fatalf("unexpected wall offer %#v error %v", result, err)
	}
}

// TestOfferRejectsProtocolOverflow verifies packet range validation.
func TestOfferRejectsProtocolOverflow(t *testing.T) {
	item := catalogmodel.Item{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: math.MaxInt64}}}
	_, err := Offer(item, furnituremodel.Definition{Kind: furnituremodel.KindFloor})
	if !errors.Is(err, ErrProtocolRange) {
		t.Fatalf("expected protocol range error, got %v", err)
	}
}
