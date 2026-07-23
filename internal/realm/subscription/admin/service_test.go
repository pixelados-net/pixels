package admin

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/realm/subscription/record"
)

// targetedStore records one validated targeted offer.
type targetedStore struct {
	Store
	// offer stores the last persisted offer.
	offer record.TargetedOffer
}

// UpsertTargetedOffer records one targeted offer.
func (store *targetedStore) UpsertTargetedOffer(_ context.Context, offer record.TargetedOffer) (record.TargetedOffer, error) {
	store.offer = offer
	return offer, nil
}

// TestSaveTargetedOfferRequiresRenderableFutureCampaign verifies Nitro-facing fields.
func TestSaveTargetedOfferRequiresRenderableFutureCampaign(t *testing.T) {
	store := &targetedStore{}
	service := New(store, nil)
	future := time.Now().Add(time.Hour)
	valid := record.TargetedOffer{CatalogItemID: 3, PurchaseLimit: 1, TitleKey: "offer.title",
		DescriptionKey: "offer.description", ImageURL: "targetedoffers/banner.png",
		IconURL: "targetedoffers/icon.png", ExpiresAt: &future, Enabled: true}
	if _, err := service.SaveTargetedOffer(context.Background(), valid); err != nil || store.offer.CatalogItemID != 3 {
		t.Fatalf("offer=%#v error=%v", store.offer, err)
	}
	invalid := valid
	invalid.ImageURL = ""
	if _, err := service.SaveTargetedOffer(context.Background(), invalid); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected missing image rejection, got %v", err)
	}
	past := time.Now().Add(-time.Hour)
	invalid, invalid.ExpiresAt = valid, &past
	if _, err := service.SaveTargetedOffer(context.Background(), invalid); !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected expired campaign rejection, got %v", err)
	}
}
