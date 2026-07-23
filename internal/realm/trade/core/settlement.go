package core

import (
	"context"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	traderecord "github.com/niflaot/pixels/internal/realm/trade/record"
	traderuntime "github.com/niflaot/pixels/internal/realm/trade/runtime"
)

const creditsType int32 = -1

// settlementItem joins an item with its definition.
type settlementItem struct {
	// item stores the inventory instance being settled.
	item furnituremodel.Item
	// definition stores the immutable transfer policy.
	definition furnituremodel.Definition
}

// settle validates both offers before transferring either side atomically.
func (service *Service) settle(ctx context.Context, session *traderuntime.Session) error {
	first, second := session.Snapshot()
	return service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		firstItems, firstCredits, err := service.validateOffer(txCtx, first)
		if err != nil {
			return err
		}
		secondItems, secondCredits, err := service.validateOffer(txCtx, second)
		if err != nil {
			return err
		}
		if err = service.transfer(txCtx, firstItems, first.PlayerID, second.PlayerID); err != nil {
			return err
		}
		if err = service.transfer(txCtx, secondItems, second.PlayerID, first.PlayerID); err != nil {
			return err
		}
		if firstCredits > 0 {
			if _, err = service.currencies.Grant(txCtx, currencyservice.GrantParams{PlayerID: second.PlayerID, CurrencyType: creditsType, Amount: firstCredits, Reason: "trade_redeemable", ActorKind: currencyservice.ActorPlayer}); err != nil {
				return err
			}
		}
		if secondCredits > 0 {
			if _, err = service.currencies.Grant(txCtx, currencyservice.GrantParams{PlayerID: first.PlayerID, CurrencyType: creditsType, Amount: secondCredits, Reason: "trade_redeemable", ActorKind: currencyservice.ActorPlayer}); err != nil {
				return err
			}
		}
		if !service.config.AuditEnabled {
			return nil
		}
		return service.store.InsertAudit(txCtx, traderecord.Audit{RoomID: session.RoomID, FirstPlayerID: first.PlayerID, SecondPlayerID: second.PlayerID, FirstIP: first.IP, SecondIP: second.IP, FirstItemIDs: first.Items, SecondItemIDs: second.Items, FirstRedeemableCredits: firstCredits, SecondRedeemableCredits: secondCredits})
	})
}

// validateOffer rechecks ownership, availability, and definition policy.
func (service *Service) validateOffer(ctx context.Context, participant traderuntime.Participant) ([]settlementItem, int64, error) {
	items := make([]settlementItem, 0, len(participant.Items))
	var credits int64
	for _, itemID := range participant.Items {
		item, found, err := service.furniture.FindItemByID(ctx, itemID)
		if err != nil {
			return nil, 0, err
		}
		if !found || item.OwnerPlayerID != participant.PlayerID || !item.InInventory() || item.MarketplaceReserved {
			return nil, 0, ErrItemUnavailable
		}
		definition, found, err := service.furniture.FindDefinitionByID(ctx, item.DefinitionID)
		if err != nil {
			return nil, 0, err
		}
		if !found || !definition.AllowTrade {
			return nil, 0, ErrItemUnavailable
		}
		items = append(items, settlementItem{item: item, definition: definition})
		credits += int64(definition.RedeemableCredits)
	}
	return items, credits, nil
}

// transfer moves regular furniture and consumes credit furniture.
func (service *Service) transfer(ctx context.Context, items []settlementItem, fromID int64, toID int64) error {
	for _, entry := range items {
		var err error
		if entry.definition.RedeemableCredits > 0 {
			err = service.furniture.DeleteInventoryItem(ctx, entry.item.ID, fromID)
		} else {
			err = service.furniture.TransferInventoryItem(ctx, entry.item.ID, fromID, toID)
		}
		if err != nil {
			return err
		}
	}
	return nil
}
