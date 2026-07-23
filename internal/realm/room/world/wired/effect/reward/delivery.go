package reward

import (
	"context"
	"fmt"
	"strconv"

	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
	rewardresult "github.com/niflaot/pixels/networking/outbound/furniture/wired/reward/result"
	inventoryrefresh "github.com/niflaot/pixels/networking/outbound/inventory/furniture/refresh"
)

// deliver applies one selected reward inside the claim transaction.
func (service *Service) deliver(ctx context.Context, node *configuration.Node, event trigger.Event, reward record.Reward) error {
	reference, err := strconv.ParseInt(reward.Reference, 10, 64)
	switch reward.Kind {
	case "badge":
		_, err = service.achievements.GrantBadge(ctx, event.PlayerID, reward.Reference, "wired")
	case "furniture":
		if err == nil {
			_, err = service.furniture.Grant(ctx, furnitureservice.GrantParams{DefinitionID: reference, OwnerPlayerID: event.PlayerID, Quantity: int32(reward.Amount)})
		}
	case "catalog_offer":
		if err == nil {
			_, err = service.catalog.Purchase(ctx, catalogservice.PurchaseParams{PlayerID: event.PlayerID, CatalogItemID: reference, HasClub: true, Amount: int32(reward.Amount), Free: true})
		}
	case "credits":
		_, err = service.currencies.Grant(ctx, currencyservice.GrantParams{PlayerID: event.PlayerID, CurrencyType: 0, Amount: reward.Amount, Reason: "wired_reward", ActorKind: currencyservice.ActorSystem})
	case "currency":
		if err == nil {
			_, err = service.currencies.Grant(ctx, currencyservice.GrantParams{PlayerID: event.PlayerID, CurrencyType: int32(reference), Amount: reward.Amount, Reason: "wired_reward", ActorKind: currencyservice.ActorSystem})
		}
	case "respect":
		key := fmt.Sprintf("wired-reward:%d:%d:%d", node.ItemID, event.ID, event.PlayerID)
		_, err = service.achievements.GrantRespect(ctx, event.PlayerID, int32(reward.Amount), key)
	default:
		return fmt.Errorf("unsupported reward kind %s", reward.Kind)
	}
	return err
}

// send projects the committed claim result and optional inventory refresh.
func (service *Service) send(ctx context.Context, playerID int64, reason int32, refresh bool) error {
	player, found := service.players.Find(playerID)
	if !found || service.connections == nil {
		return nil
	}
	peer := player.Peer()
	connection, found := service.connections.Get(peer.ConnectionKind(), peer.ConnectionID())
	if !found {
		return nil
	}
	packet, err := rewardresult.Encode(reason)
	if err != nil {
		return err
	}
	if err = connection.Send(ctx, packet); err != nil || !refresh {
		return err
	}
	packet, err = inventoryrefresh.Encode()
	if err != nil {
		return err
	}
	return connection.Send(ctx, packet)
}
