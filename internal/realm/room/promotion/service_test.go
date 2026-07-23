package promotion

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/permission"
	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// promotionStoreFixture records transaction and promotion mutations.
type promotionStoreFixture struct {
	Store
	upserts        int
	transactionErr error
	value          Promotion
}

// WithinTransaction executes work and optionally simulates commit failure.
func (store *promotionStoreFixture) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	if err := work(ctx); err != nil {
		return err
	}
	return store.transactionErr
}

// Upsert records one promotion mutation.
func (store *promotionStoreFixture) Upsert(context.Context, PurchaseParams, Config) (Promotion, error) {
	store.upserts++
	return store.value, nil
}

// roomManagerFixture exposes one owned room.
type roomManagerFixture struct {
	roomservice.Manager
	room roommodel.Room
}

// FindByID returns the fixture room.
func (manager roomManagerFixture) FindByID(context.Context, int64) (roommodel.Room, bool, error) {
	return manager.room, true, nil
}

// ListByOwner returns the fixture room.
func (manager roomManagerFixture) ListByOwner(context.Context, int64) ([]roommodel.Room, error) {
	return []roommodel.Room{manager.room}, nil
}

// promotionCatalogFixture exposes one service offer and deferred projection.
type promotionCatalogFixture struct {
	catalogservice.Manager
	completed bool
}

// Page returns one Room Ads page and offer.
func (catalog *promotionCatalogFixture) Page(context.Context, int64, int64, bool) (catalogmodel.Page, []catalogmodel.Item, error) {
	return catalogmodel.Page{Layout: "room_ads"}, []catalogmodel.Item{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 990001}}}}, nil
}

// PurchaseWithin records post-commit projection without external state.
func (catalog *promotionCatalogFixture) PurchaseWithin(context.Context, catalogservice.PurchaseParams) (catalogservice.PurchaseResult, func(context.Context), error) {
	return catalogservice.PurchaseResult{}, func(context.Context) { catalog.completed = true }, nil
}

// TestPurchaseCommitsChargeAndPromotionBeforeProjection verifies the shared transaction boundary.
func TestPurchaseCommitsChargeAndPromotionBeforeProjection(t *testing.T) {
	now := time.Unix(100, 0)
	store := &promotionStoreFixture{value: Promotion{ID: 5, RoomID: 160, StartsAt: now, EndsAt: now.Add(time.Hour)}}
	catalog := &promotionCatalogFixture{}
	rooms := roomManagerFixture{room: roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 160}}, OwnerPlayerID: 1}}
	service := New(Config{}, store, rooms, catalog, nil, permission.Node(""), nil, nil)
	service.now = func() time.Time { return now }
	value, err := service.Purchase(context.Background(), PurchaseParams{PlayerID: 1, RoomID: 160, PageID: 990, OfferID: 990001, Title: "QA"})
	if err != nil || value.ID != 5 || store.upserts != 1 || !catalog.completed {
		t.Fatalf("value=%+v upserts=%d completed=%t err=%v", value, store.upserts, catalog.completed, err)
	}
}

// TestPurchaseDoesNotProjectAfterCommitFailure verifies failed transactions remain externally silent.
func TestPurchaseDoesNotProjectAfterCommitFailure(t *testing.T) {
	expected := errors.New("commit failed")
	store := &promotionStoreFixture{transactionErr: expected, value: Promotion{ID: 5, RoomID: 160}}
	catalog := &promotionCatalogFixture{}
	rooms := roomManagerFixture{room: roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 160}}, OwnerPlayerID: 1}}
	service := New(Config{}, store, rooms, catalog, nil, permission.Node(""), nil, nil)
	_, err := service.Purchase(context.Background(), PurchaseParams{PlayerID: 1, RoomID: 160, PageID: 990, OfferID: 990001, Title: "QA"})
	if !errors.Is(err, expected) || catalog.completed {
		t.Fatalf("completed=%t err=%v", catalog.completed, err)
	}
}

// TestPurchaseRejectsNonOwnerBeforeCharging verifies room authorization precedes commerce.
func TestPurchaseRejectsNonOwnerBeforeCharging(t *testing.T) {
	store := &promotionStoreFixture{}
	catalog := &promotionCatalogFixture{}
	rooms := roomManagerFixture{room: roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 160}}, OwnerPlayerID: 2}}
	service := New(Config{}, store, rooms, catalog, nil, permission.Node(""), nil, nil)
	_, err := service.Purchase(context.Background(), PurchaseParams{PlayerID: 1, RoomID: 160, PageID: 990, OfferID: 990001, Title: "QA"})
	if !errors.Is(err, ErrNoRights) || store.upserts != 0 || catalog.completed {
		t.Fatalf("upserts=%d completed=%t err=%v", store.upserts, catalog.completed, err)
	}
}
