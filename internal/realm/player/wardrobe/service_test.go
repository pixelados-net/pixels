package wardrobe

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	playerfigure "github.com/niflaot/pixels/internal/realm/player/figure"
)

// wardrobeStore records focused wardrobe mutations.
type wardrobeStore struct{ outfit Outfit }

// Outfits returns the stored outfit.
func (store *wardrobeStore) Outfits(context.Context, int64) ([]Outfit, error) {
	return []Outfit{store.outfit}, nil
}

// SaveOutfit records one outfit.
func (store *wardrobeStore) SaveOutfit(_ context.Context, _ int64, outfit Outfit) error {
	store.outfit = outfit
	return nil
}

// Clothing returns an empty unlock snapshot.
func (*wardrobeStore) Clothing(context.Context, int64) (ClothingSnapshot, error) {
	return ClothingSnapshot{}, nil
}

// RedeemClothing returns one deterministic unlock.
func (*wardrobeStore) RedeemClothing(context.Context, int64, int64) (RedeemResult, error) {
	return RedeemResult{Applied: true, Snapshot: ClothingSnapshot{FigureSetIDs: []int32{3356}}}, nil
}

// TestSaveValidatesAndNormalizesOutfit verifies slot and gender policy.
func TestSaveValidatesAndNormalizesOutfit(t *testing.T) {
	store := &wardrobeStore{}
	service := New(store)
	if err := service.Save(context.Background(), 1, Outfit{SlotID: 1, Figure: " hd-180-1 ", Gender: "m"}); err != nil {
		t.Fatalf("save outfit: %v", err)
	}
	if store.outfit.Gender != "M" || store.outfit.Figure != "hd-180-1" {
		t.Fatalf("unexpected outfit %#v", store.outfit)
	}
	if err := service.Save(context.Background(), 1, Outfit{SlotID: 0, Figure: "x", Gender: "M"}); !errors.Is(err, ErrInvalidOutfit) {
		t.Fatalf("expected invalid slot, got %v", err)
	}
}

// TestRedeemValidatesIdentity verifies invalid item requests do not reach persistence.
func TestRedeemValidatesIdentity(t *testing.T) {
	if _, err := New(&wardrobeStore{}).Redeem(context.Background(), 1, 0); !errors.Is(err, ErrInvalidClothingItem) {
		t.Fatalf("expected invalid item, got %v", err)
	}
}

// TestWardrobeReadsAndRedeems verifies cold-path snapshots and successful redemption.
func TestWardrobeReadsAndRedeems(t *testing.T) {
	store := &wardrobeStore{outfit: Outfit{SlotID: 2, Figure: "hd-180-1", Gender: "M"}}
	service := New(store)
	outfits, err := service.Outfits(context.Background(), 1)
	if err != nil || len(outfits) != 1 || outfits[0].SlotID != 2 {
		t.Fatalf("outfits=%#v err=%v", outfits, err)
	}
	if _, err = service.Clothing(context.Background(), 0); !errors.Is(err, ErrInvalidClothingItem) {
		t.Fatalf("expected invalid clothing owner, got %v", err)
	}
	clothing, err := service.Clothing(context.Background(), 1)
	if err != nil || len(clothing.FigureSetIDs) != 0 {
		t.Fatalf("clothing=%#v err=%v", clothing, err)
	}
	result, err := service.Redeem(context.Background(), 1, 9)
	if err != nil || !result.Applied || len(result.Snapshot.FigureSetIDs) != 1 {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

// TestWardrobeFigureEntitlementAndConfig verifies authoritative figuredata and policy bounds.
func TestWardrobeFigureEntitlementAndConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "figuredata.xml")
	data := []byte(`<figuredata><colors><palette id="1"><color id="1" club="0" selectable="1"/></palette></colors><sets><settype type="hd" paletteid="1"><set id="180" gender="U" club="0" selectable="1"/></settype></sets></figuredata>`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	catalog, err := playerfigure.NewCatalog(playerfigure.Config{Path: path})
	if err != nil {
		t.Fatal(err)
	}
	store := &wardrobeStore{}
	service := NewConfigured(store, catalog, nil, Config{})
	if err = service.Save(context.Background(), 1, Outfit{SlotID: 1, Figure: "hd-180-1", Gender: "M"}); err != nil {
		t.Fatalf("save entitled outfit: %v", err)
	}
	if err = service.Save(context.Background(), 1, Outfit{SlotID: 1, Figure: "hd-999-1", Gender: "M"}); !errors.Is(err, ErrInvalidOutfit) {
		t.Fatalf("expected unknown figure rejection, got %v", err)
	}
	t.Setenv("PIXELS_PLAYER_WARDROBE_MAX_SLOT", "8")
	config, err := LoadConfig()
	if err != nil || config.MaximumSlot != 8 {
		t.Fatalf("config=%#v err=%v", config, err)
	}
}
