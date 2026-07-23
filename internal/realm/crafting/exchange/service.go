// Package exchange owns furniture-for-credit redemption.
package exchange

import (
	"context"

	redeemedevent "github.com/niflaot/pixels/internal/realm/crafting/exchange/events/redeemed"
	craftingrecord "github.com/niflaot/pixels/internal/realm/crafting/record"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	"github.com/niflaot/pixels/pkg/bus"
)

const creditsType int32 = -1

// Service coordinates atomic furniture exchange redemption.
type Service struct {
	// store starts the shared transaction.
	store craftingrecord.Store
	// furniture validates and consumes exchange furniture.
	furniture Furniture
	// currencies credits the durable wallet.
	currencies currencyservice.Granter
	// events publishes committed exchange outcomes.
	events bus.Publisher
}

// Result stores one committed exchange projection.
type Result struct {
	// RemovedItemID identifies the consumed furniture instance.
	RemovedItemID int64
	// Credits stores the exact granted credit amount.
	Credits int64
	// Item stores the consumed record for room projection.
	Item furnituremodel.Item
}

// Furniture exposes the guarded furniture operations exchange consumes.
type Furniture interface {
	furnitureservice.TradingManager
	Pickup(context.Context, furnitureservice.PickupParams) (furnituremodel.Item, error)
}

// New creates an item exchange service.
func New(store craftingrecord.Store, furniture *furnitureservice.Service, currencies currencyservice.Granter, events bus.Publisher) *Service {
	return &Service{store: store, furniture: furniture, currencies: currencies, events: events}
}

// Redeem destroys one owned inventory item and credits its declared value.
func (service *Service) Redeem(ctx context.Context, playerID int64, itemID int64) (Result, error) {
	result := Result{RemovedItemID: itemID}
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		item, found, err := service.furniture.FindItemByID(txCtx, itemID)
		if err != nil {
			return err
		}
		if !found || item.OwnerPlayerID != playerID || item.MarketplaceReserved {
			return craftingrecord.ErrItemUnavailable
		}
		result.Item = item
		definition, found, err := service.furniture.FindDefinitionByID(txCtx, item.DefinitionID)
		if err != nil {
			return err
		}
		if !found || definition.RedeemableCredits <= 0 {
			return craftingrecord.ErrExchangeValue
		}
		if item.InRoom() {
			_, err = service.furniture.Pickup(txCtx, furnitureservice.PickupParams{ItemID: itemID, ActorPlayerID: playerID, RoomID: *item.RoomID})
			if err != nil {
				return craftingrecord.ErrItemUnavailable
			}
		}
		if err = service.furniture.DeleteInventoryItem(txCtx, itemID, playerID); err != nil {
			return craftingrecord.ErrItemUnavailable
		}
		result.Credits = int64(definition.RedeemableCredits)
		_, err = service.currencies.Grant(txCtx, currencyservice.GrantParams{PlayerID: playerID, CurrencyType: creditsType, Amount: result.Credits, Reason: "item_exchange", ActorKind: currencyservice.ActorPlayer, ActorID: &playerID})
		return err
	})
	if err == nil && service.events != nil {
		_ = service.events.Publish(context.Background(), bus.Event{Name: redeemedevent.Name, Payload: redeemedevent.Payload{PlayerID: playerID, ItemID: itemID, Credits: result.Credits}})
	}
	return result, err
}
