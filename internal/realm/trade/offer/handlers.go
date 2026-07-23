// Package offer adapts direct-trade furniture offer packets.
package offer

import (
	"context"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	tradecore "github.com/niflaot/pixels/internal/realm/trade/core"
	traderuntime "github.com/niflaot/pixels/internal/realm/trade/runtime"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inadd "github.com/niflaot/pixels/networking/inbound/trade/additem"
	inadds "github.com/niflaot/pixels/networking/inbound/trade/additems"
	inremove "github.com/niflaot/pixels/networking/inbound/trade/removeitem"
	outnosuchitem "github.com/niflaot/pixels/networking/outbound/trade/nosuchitem"
	outnotopen "github.com/niflaot/pixels/networking/outbound/trade/notopen"
	outupdate "github.com/niflaot/pixels/networking/outbound/trade/update"
)

// Handler owns trade offer adapters.
type Handler struct {
	// Service executes direct-trade offer behavior.
	Service *tradecore.Service
	// Sender projects packets to both participants.
	Sender *traderuntime.Sender
	// Furniture reads item definitions for protocol projection.
	Furniture furnitureservice.TradingManager
}

// Register installs direct-trade offer handlers.
func Register(registry *netconn.HandlerRegistry, handler Handler) {
	_ = registry.Register(inadd.Header, handler.addOne)
	_ = registry.Register(inadds.Header, handler.addMany)
	_ = registry.Register(inremove.Header, handler.remove)
}

// addOne stages one item.
func (handler Handler) addOne(connection netconn.Context, packet codec.Packet) error {
	itemID, err := inadd.Decode(packet)
	if err != nil {
		return err
	}
	return handler.mutate(connection, []int64{itemID}, 0)
}

// addMany stages several items.
func (handler Handler) addMany(connection netconn.Context, packet codec.Packet) error {
	itemIDs, err := inadds.Decode(packet)
	if err != nil {
		return err
	}
	return handler.mutate(connection, itemIDs, 0)
}

// remove unstages one item.
func (handler Handler) remove(connection netconn.Context, packet codec.Packet) error {
	itemID, err := inremove.Decode(packet)
	if err != nil {
		return err
	}
	return handler.mutate(connection, nil, itemID)
}

// mutate changes an offer and broadcasts its full projection.
func (handler Handler) mutate(connection netconn.Context, add []int64, remove int64) error {
	playerID, err := handler.Sender.PlayerID(connection)
	if err != nil {
		return err
	}
	session, found := handler.Service.Registry().Find(playerID)
	if !found {
		response, _ := outnotopen.Encode()
		return handler.Sender.Send(context.Background(), playerID, response)
	}
	if len(add) > 0 {
		err = handler.Service.Offer(context.Background(), playerID, add)
	} else {
		err = handler.Service.Remove(playerID, remove)
	}
	if err != nil {
		response, _ := outnosuchitem.Encode()
		return handler.Sender.Send(context.Background(), playerID, response)
	}
	response, err := handler.updatePacket(context.Background(), session)
	if err != nil {
		return err
	}
	return handler.Sender.Both(context.Background(), session, response)
}

// updatePacket builds one full trade-offer projection.
func (handler Handler) updatePacket(ctx context.Context, session *traderuntime.Session) (codec.Packet, error) {
	first, second := session.Snapshot()
	firstView, err := handler.participant(ctx, first)
	if err != nil {
		return codec.Packet{}, err
	}
	secondView, err := handler.participant(ctx, second)
	if err != nil {
		return codec.Packet{}, err
	}
	return outupdate.Encode(firstView, secondView)
}

// participant enriches offered item ids.
func (handler Handler) participant(ctx context.Context, participant traderuntime.Participant) (outupdate.Participant, error) {
	result := outupdate.Participant{PlayerID: participant.PlayerID, Items: make([]outupdate.Item, 0, len(participant.Items))}
	for _, itemID := range participant.Items {
		item, found, err := handler.Furniture.FindItemByID(ctx, itemID)
		if err != nil {
			return outupdate.Participant{}, err
		}
		if !found {
			continue
		}
		definition, found, err := handler.Furniture.FindDefinitionByID(ctx, item.DefinitionID)
		if err != nil {
			return outupdate.Participant{}, err
		}
		if !found {
			continue
		}
		result.Items = append(result.Items, outupdate.Item{Item: item, Definition: definition})
		result.RedeemableCredits += int64(definition.RedeemableCredits)
	}
	return result, nil
}
