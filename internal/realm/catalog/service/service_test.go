package service

import (
	"context"
	"sync"
	"testing"

	"github.com/niflaot/pixels/internal/realm/catalog/model"
	"github.com/niflaot/pixels/internal/realm/catalog/repository"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
	"go.uber.org/zap"
)

// fakeStore contains catalog service persistence fixtures.
type fakeStore struct {
	// Store supplies methods outside each focused test's behavior.
	repository.Store

	// mutex serializes transaction work and snapshot reads.
	mutex sync.Mutex

	// pages stores persistent page fixtures.
	pages []model.Page

	// items stores persistent offer fixtures.
	items []model.Item

	// sanitize stores orphan definition fixtures.
	sanitize []furnituremodel.Definition

	// available reports whether one LTD unit remains.
	available bool

	// reserved reports whether the LTD unit is currently reserved.
	reserved bool

	// txCalls counts opened transactions.
	txCalls int
}

// ListPages lists fixture pages.
func (store *fakeStore) ListPages(context.Context) ([]model.Page, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	return append([]model.Page{}, store.pages...), nil
}

// ListItems lists fixture offers.
func (store *fakeStore) ListItems(context.Context, *int64) ([]model.Item, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	return append([]model.Item{}, store.items...), nil
}

// SanitizeList lists fixture orphan definitions.
func (store *fakeStore) SanitizeList(context.Context) ([]furnituremodel.Definition, error) {
	return append([]furnituremodel.Definition{}, store.sanitize...), nil
}

// WithinTransaction serializes work and restores catalog state on failure.
func (store *fakeStore) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	store.txCalls++
	reserved := store.reserved
	items := append([]model.Item{}, store.items...)
	err := work(ctx)
	if err != nil {
		store.reserved = reserved
		store.items = items
	}

	return err
}

// ReserveLimitedUnit reserves one fixture LTD unit.
func (store *fakeStore) ReserveLimitedUnit(context.Context, int64, int64) (int32, bool, error) {
	if !store.available || store.reserved {
		return 0, false, nil
	}
	store.reserved = true

	return 1, true, nil
}

// CompleteLimitedUnit completes one fixture LTD unit.
func (store *fakeStore) CompleteLimitedUnit(context.Context, int64, int32, int64, int64) (bool, error) {
	if !store.reserved {
		return false, nil
	}
	store.available = false
	store.items[0].LimitedSells++
	store.items[0].Enabled = false

	return true, nil
}

// fakeCurrency records catalog charges.
type fakeCurrency struct {
	// mutex protects concurrent charge state.
	mutex sync.Mutex

	// balance stores the current fake balance.
	balance int64

	// calls stores committed grant inputs.
	calls []currencyservice.GrantParams

	// err stores an optional charge failure.
	err error
}

// Grant applies one fake signed currency change.
func (currency *fakeCurrency) Grant(_ context.Context, params currencyservice.GrantParams) (int64, error) {
	currency.mutex.Lock()
	defer currency.mutex.Unlock()
	if currency.err != nil {
		return 0, currency.err
	}
	currency.balance += params.Amount
	currency.calls = append(currency.calls, params)

	return currency.balance, nil
}

// fakeFurniture records catalog furniture grants.
type fakeFurniture struct {
	// mutex protects concurrent grant state.
	mutex sync.Mutex

	// calls stores committed grant inputs.
	calls []furnitureservice.GrantParams
	// giftCalls stores committed wrapped grant inputs.
	giftCalls []furnitureservice.GiftGrantParams

	// err stores an optional grant failure.
	err error

	// definitions stores furniture metadata fixtures.
	definitions []furnituremodel.Definition
}

// fakeTeleportPairer records teleport relationships created by purchases.
type fakeTeleportPairer struct {
	// pairs stores paired item ids in call order.
	pairs [][2]int64
	// err stores an optional pairing failure.
	err error
}

// PairTeleports records one requested teleport relationship.
func (pairer *fakeTeleportPairer) PairTeleports(_ context.Context, _ int64, firstItemID int64, secondItemID int64) error {
	if pairer.err != nil {
		return pairer.err
	}
	pairer.pairs = append(pairer.pairs, [2]int64{firstItemID, secondItemID})

	return nil
}

// FindDefinitionByID finds one furniture metadata fixture.
func (furniture *fakeFurniture) FindDefinitionByID(_ context.Context, id int64) (furnituremodel.Definition, bool, error) {
	for _, definition := range furniture.definitions {
		if definition.ID == id {
			return definition, true, nil
		}
	}

	return furnituremodel.Definition{}, false, nil
}

// ListDefinitions lists furniture metadata fixtures.
func (furniture *fakeFurniture) ListDefinitions(context.Context) ([]furnituremodel.Definition, error) {
	return append([]furnituremodel.Definition{}, furniture.definitions...), nil
}

// Grant creates fake inventory items.
func (furniture *fakeFurniture) Grant(_ context.Context, params furnitureservice.GrantParams) ([]furnituremodel.Item, error) {
	furniture.mutex.Lock()
	defer furniture.mutex.Unlock()
	if furniture.err != nil {
		return nil, furniture.err
	}
	furniture.calls = append(furniture.calls, params)
	items := make([]furnituremodel.Item, 0, params.Quantity)
	for index := int32(0); index < params.Quantity; index++ {
		items = append(items, furnituremodel.Item{
			Base:         sharedmodel.Base{Identity: sharedmodel.Identity{ID: int64(index + 20)}},
			DefinitionID: params.DefinitionID, OwnerPlayerID: params.OwnerPlayerID,
		})
	}

	return items, nil
}

// GrantGift creates fake wrapped inventory items.
func (furniture *fakeFurniture) GrantGift(ctx context.Context, params furnitureservice.GiftGrantParams) ([]furnituremodel.Item, error) {
	furniture.mutex.Lock()
	furniture.giftCalls = append(furniture.giftCalls, params)
	furniture.mutex.Unlock()

	return furniture.Grant(ctx, params.GrantParams)
}

// serviceFixture contains catalog service test collaborators.
type serviceFixture struct {
	// service stores the tested catalog behavior.
	service *Service

	// store stores fake catalog persistence.
	store *fakeStore

	// currency stores fake balance behavior.
	currency *fakeCurrency

	// furniture stores fake grant behavior.
	furniture *fakeFurniture

	// teleportPairs stores fake teleport pairing behavior.
	teleportPairs *fakeTeleportPairer
}

// newServiceFixture creates and refreshes catalog test behavior.
func newServiceFixture(t *testing.T, item model.Item) serviceFixture {
	t.Helper()
	store := &fakeStore{pages: []model.Page{pageForTest()}, items: []model.Item{item}, available: true}
	currency := &fakeCurrency{balance: 100}
	furniture := &fakeFurniture{definitions: []furnituremodel.Definition{{
		Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: item.DefinitionID}}, SpriteID: 1, Name: item.Name,
	}}}
	teleportPairs := &fakeTeleportPairer{}
	service := New(store, currency, furniture, nil, zap.NewNop()).WithTeleportPairer(teleportPairs)
	if err := service.Refresh(context.Background()); err != nil {
		t.Fatalf("refresh fixture: %v", err)
	}

	return serviceFixture{service: service, store: store, currency: currency, furniture: furniture, teleportPairs: teleportPairs}
}

// pageForTest returns one accessible catalog page.
func pageForTest() model.Page {
	return model.Page{
		Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, Name: "chairs",
		Layout: model.DefaultLayout, Visible: true, Enabled: true,
	}
}

// itemForTest returns one enabled credits offer.
func itemForTest() model.Item {
	return model.Item{
		Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 10}}, PageID: 1,
		DefinitionID: 2, Name: "chair_plasto", CostCredits: 10, PointsType: model.CreditsType,
		Amount: 1, Enabled: true, ExtraData: "0",
	}
}
