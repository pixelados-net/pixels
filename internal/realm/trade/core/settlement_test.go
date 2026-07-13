package core

import (
	"context"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	traderecord "github.com/niflaot/pixels/internal/realm/trade/record"
	traderuntime "github.com/niflaot/pixels/internal/realm/trade/runtime"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
	"testing"
)

// settlementStore executes work without persistence for focused settlement tests.
type settlementStore struct{}

func (*settlementStore) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	return work(ctx)
}
func (*settlementStore) InsertAudit(context.Context, traderecord.Audit) error { return nil }
func (*settlementStore) ListAudits(context.Context, int64, int32) ([]traderecord.Audit, error) {
	return nil, nil
}

// settlementFurniture provides stable owned inventory items.
type settlementFurniture struct{}

func (*settlementFurniture) FindDefinitionByID(context.Context, int64) (furnituremodel.Definition, bool, error) {
	return furnituremodel.Definition{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, AllowTrade: true}, true, nil
}
func (*settlementFurniture) ListDefinitions(context.Context) ([]furnituremodel.Definition, error) {
	return nil, nil
}
func (*settlementFurniture) FindItemByID(_ context.Context, itemID int64) (furnituremodel.Item, bool, error) {
	owner := int64(1)
	if itemID == 20 {
		owner = 2
	}
	return furnituremodel.Item{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: itemID}}, DefinitionID: 1, OwnerPlayerID: owner}, true, nil
}
func (*settlementFurniture) ListInventory(context.Context, int64) ([]furnituremodel.Item, error) {
	return nil, nil
}
func (*settlementFurniture) ListRoomItems(context.Context, int64) ([]furnituremodel.Item, error) {
	return nil, nil
}
func (*settlementFurniture) ReserveForMarketplace(context.Context, int64, int64) (furnituremodel.Item, furnituremodel.Definition, error) {
	return furnituremodel.Item{}, furnituremodel.Definition{}, nil
}
func (*settlementFurniture) ReleaseFromMarketplace(context.Context, int64, int64) error { return nil }
func (*settlementFurniture) TransferFromMarketplace(context.Context, int64, int64, int64) error {
	return nil
}
func (*settlementFurniture) TransferInventoryItem(context.Context, int64, int64, int64) error {
	return nil
}
func (*settlementFurniture) DeleteInventoryItem(context.Context, int64, int64) error { return nil }

// settlementCurrency accepts settlement grants.
type settlementCurrency struct{}

func (*settlementCurrency) Grant(context.Context, currencyservice.GrantParams) (int64, error) {
	return 0, nil
}

// TestSettleRevalidatesAndTransfersBothOffers verifies the happy settlement path.
func TestSettleRevalidatesAndTransfersBothOffers(t *testing.T) {
	service := &Service{config: Options{}, store: &settlementStore{}, furniture: &settlementFurniture{}, currencies: &settlementCurrency{}}
	session := &traderuntime.Session{RoomID: 9, First: traderuntime.Participant{PlayerID: 1, Items: []int64{10}}, Second: traderuntime.Participant{PlayerID: 2, Items: []int64{20}}}
	if err := service.settle(context.Background(), session); err != nil {
		t.Fatal(err)
	}
}

// BenchmarkSettle measures validation and atomic ownership settlement coordination.
func BenchmarkSettle(b *testing.B) {
	service := &Service{config: Options{}, store: &settlementStore{}, furniture: &settlementFurniture{}, currencies: &settlementCurrency{}}
	session := &traderuntime.Session{RoomID: 9, First: traderuntime.Participant{PlayerID: 1, Items: []int64{10}}, Second: traderuntime.Participant{PlayerID: 2, Items: []int64{20}}}
	b.ReportAllocs()
	for b.Loop() {
		if err := service.settle(context.Background(), session); err != nil {
			b.Fatal(err)
		}
	}
}
