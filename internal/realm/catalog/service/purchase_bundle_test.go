package service

import (
	"context"
	"errors"
	"testing"

	catalogpurchased "github.com/niflaot/pixels/internal/realm/catalog/events/purchased"
	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
	roombundle "github.com/niflaot/pixels/internal/realm/room/record/bundle"
	roompurchased "github.com/niflaot/pixels/internal/realm/room/record/events/bundlepurchased"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	"github.com/niflaot/pixels/pkg/bus"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// purchaseEventPublisher records completed purchase events.
type purchaseEventPublisher struct {
	// events stores published events in order.
	events []bus.Event
}

// Publish records one event.
func (publisher *purchaseEventPublisher) Publish(_ context.Context, event bus.Event) error {
	publisher.events = append(publisher.events, event)
	return nil
}

// purchaseRoomBundles stores room bundle test behavior.
type purchaseRoomBundles struct {
	// result stores the configured clone result.
	result roombundle.CloneResult
	// err stores the configured clone error.
	err error
	// calls stores clone invocation count.
	calls int
}

// Clone returns configured room creation behavior.
func (rooms *purchaseRoomBundles) Clone(context.Context, roombundle.CloneParams) (roombundle.CloneResult, error) {
	rooms.calls++
	return rooms.result, rooms.err
}

// Preview returns one grouped template product for purchase tests.
func (*purchaseRoomBundles) Preview(context.Context, int64) ([]roombundle.Product, error) {
	return []roombundle.Product{{DefinitionID: 1, Quantity: 2}}, nil
}

// Mark is unused by purchase tests.
func (*purchaseRoomBundles) Mark(context.Context, int64) (roommodel.Room, error) {
	return roommodel.Room{}, nil
}

// Unmark is unused by purchase tests.
func (*purchaseRoomBundles) Unmark(context.Context, int64) (roommodel.Room, error) {
	return roommodel.Room{}, nil
}

// Templates is unused by purchase tests.
func (*purchaseRoomBundles) Templates(context.Context) ([]roommodel.Room, error) { return nil, nil }

// FindTemplate is unused by purchase tests.
func (*purchaseRoomBundles) FindTemplate(context.Context, int64) (roommodel.Room, bool, error) {
	return roommodel.Room{}, false, nil
}

// purchasePlayerFinder resolves one buyer.
type purchasePlayerFinder struct{}

// FindByID resolves the bundle buyer.
func (purchasePlayerFinder) FindByID(context.Context, int64) (playerservice.Record, bool, error) {
	return playerservice.Record{Player: playermodel.Player{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 7}}, Username: "demo"}}, true, nil
}

// FindByUsername is unused by bundle purchases.
func (purchasePlayerFinder) FindByUsername(context.Context, string) (playerservice.Record, bool, error) {
	return playerservice.Record{}, false, nil
}

// roomBundleItemForTest returns one enabled room bundle offer.
func roomBundleItemForTest() catalogmodel.Item {
	templateID := int64(100)
	return catalogmodel.Item{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1101}}, PageID: 1, RoomBundleTemplateRoomID: &templateID, Name: "starter_loft_bundle", CostCredits: 75, PointsType: catalogmodel.CreditsType, Enabled: true}
}

// TestPurchaseRoomBundleCreatesRoomAndCharges verifies the specialized grant branch.
func TestPurchaseRoomBundleCreatesRoomAndCharges(t *testing.T) {
	fixture := newServiceFixture(t, roomBundleItemForTest())
	events := &purchaseEventPublisher{}
	fixture.service.events = events
	rooms := &purchaseRoomBundles{result: roombundle.CloneResult{Room: roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 44}}, Name: "Starter Loft"}, FurnitureCount: 7}}
	fixture.service.WithPlayers(purchasePlayerFinder{}).WithRoomBundles(rooms)
	result, err := fixture.service.Purchase(context.Background(), PurchaseParams{PlayerID: 7, CatalogItemID: 1101, Amount: 1})
	if err != nil || result.CreatedRoomID == nil || *result.CreatedRoomID != 44 || result.ClonedFurnitureCount != 7 {
		t.Fatalf("result=%#v error=%v", result, err)
	}
	if rooms.calls != 1 || len(fixture.currency.calls) != 1 || len(fixture.furniture.calls) != 0 {
		t.Fatalf("room calls=%d currency=%#v furniture=%#v", rooms.calls, fixture.currency.calls, fixture.furniture.calls)
	}
	if len(events.events) != 2 || events.events[0].Name != catalogpurchased.Name || events.events[1].Name != roompurchased.Name {
		t.Fatalf("events=%#v", events.events)
	}
	payload, ok := events.events[0].Payload.(catalogpurchased.Payload)
	if !ok || payload.CreatedRoomID == nil || *payload.CreatedRoomID != 44 {
		t.Fatalf("catalog payload=%#v", events.events[0].Payload)
	}
}

// TestPurchaseRoomBundleLimitDoesNotCharge verifies room capacity rejects before currency mutation.
func TestPurchaseRoomBundleLimitDoesNotCharge(t *testing.T) {
	fixture := newServiceFixture(t, roomBundleItemForTest())
	rooms := &purchaseRoomBundles{err: roombundle.ErrRoomLimitReached}
	fixture.service.WithPlayers(purchasePlayerFinder{}).WithRoomBundles(rooms)
	_, err := fixture.service.Purchase(context.Background(), PurchaseParams{PlayerID: 7, CatalogItemID: 1101, Amount: 1})
	if !errors.Is(err, roombundle.ErrRoomLimitReached) || len(fixture.currency.calls) != 0 {
		t.Fatalf("currency=%#v error=%v", fixture.currency.calls, err)
	}
}

// TestPurchaseRoomBundleRejectsBulkAmount verifies quantity cannot duplicate rooms.
func TestPurchaseRoomBundleRejectsBulkAmount(t *testing.T) {
	fixture := newServiceFixture(t, roomBundleItemForTest())
	rooms := &purchaseRoomBundles{}
	fixture.service.WithPlayers(purchasePlayerFinder{}).WithRoomBundles(rooms)
	_, err := fixture.service.Purchase(context.Background(), PurchaseParams{PlayerID: 7, CatalogItemID: 1101, Amount: 2})
	if !errors.Is(err, ErrInvalidAmount) || rooms.calls != 0 {
		t.Fatalf("calls=%d error=%v", rooms.calls, err)
	}
}
