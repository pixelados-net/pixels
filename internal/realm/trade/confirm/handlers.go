// Package confirm adapts direct-trade agreement packets.
package confirm

import (
	"context"

	tradecore "github.com/niflaot/pixels/internal/realm/trade/core"
	traderuntime "github.com/niflaot/pixels/internal/realm/trade/runtime"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inaccept "github.com/niflaot/pixels/networking/inbound/trade/accept"
	inconfirm "github.com/niflaot/pixels/networking/inbound/trade/confirm"
	inunaccept "github.com/niflaot/pixels/networking/inbound/trade/unaccept"
	outrefresh "github.com/niflaot/pixels/networking/outbound/inventory/furniture/refresh"
	outaccepted "github.com/niflaot/pixels/networking/outbound/trade/accepted"
	outclosed "github.com/niflaot/pixels/networking/outbound/trade/closed"
	outcompleted "github.com/niflaot/pixels/networking/outbound/trade/completed"
	outconfirmation "github.com/niflaot/pixels/networking/outbound/trade/confirmation"
	outnotopen "github.com/niflaot/pixels/networking/outbound/trade/notopen"
)

// Handler owns trade agreement adapters.
type Handler struct {
	// Service executes direct-trade agreement and settlement behavior.
	Service *tradecore.Service
	// Sender projects packets to both participants.
	Sender *traderuntime.Sender
}

// Register installs direct-trade agreement handlers.
func Register(registry *netconn.HandlerRegistry, handler Handler) {
	_ = registry.Register(inaccept.Header, handler.accept)
	_ = registry.Register(inunaccept.Header, handler.unaccept)
	_ = registry.Register(inconfirm.Header, handler.confirm)
}

// accept applies first-phase acceptance.
func (handler Handler) accept(connection netconn.Context, packet codec.Packet) error {
	if err := inaccept.Decode(packet); err != nil {
		return err
	}
	return handler.setAccepted(connection, true)
}

// unaccept revokes first-phase acceptance.
func (handler Handler) unaccept(connection netconn.Context, packet codec.Packet) error {
	if err := inunaccept.Decode(packet); err != nil {
		return err
	}
	return handler.setAccepted(connection, false)
}

// setAccepted broadcasts agreement and opens confirmation when both accepted.
func (handler Handler) setAccepted(connection netconn.Context, value bool) error {
	playerID, err := handler.Sender.PlayerID(connection)
	if err != nil {
		return err
	}
	session, found := handler.Service.Registry().Find(playerID)
	if !found {
		response, _ := outnotopen.Encode()
		return handler.Sender.Send(context.Background(), playerID, response)
	}
	participant, _ := session.Participant(playerID)
	both, err := handler.Service.Accept(playerID, value)
	if err != nil {
		return err
	}
	response, err := outaccepted.Encode(participant.PlayerID, value)
	if err != nil {
		return err
	}
	if err = handler.Sender.Both(context.Background(), session, response); err != nil {
		return err
	}
	if both {
		confirmation, _ := outconfirmation.Encode()
		return handler.Sender.Both(context.Background(), session, confirmation)
	}
	return nil
}

// confirm applies final confirmation and broadcasts completion.
func (handler Handler) confirm(connection netconn.Context, packet codec.Packet) error {
	if err := inconfirm.Decode(packet); err != nil {
		return err
	}
	playerID, err := handler.Sender.PlayerID(connection)
	if err != nil {
		return err
	}
	session, found := handler.Service.Registry().Find(playerID)
	if !found {
		response, _ := outnotopen.Encode()
		return handler.Sender.Send(context.Background(), playerID, response)
	}
	participant, _ := session.Participant(playerID)
	completed, serviceErr := handler.Service.Confirm(context.Background(), playerID)
	if serviceErr != nil {
		response, _ := outclosed.Encode(participant.PlayerID, 1)
		return handler.Sender.Both(context.Background(), session, response)
	}
	if completed {
		packets, encodeErr := completionPackets()
		if encodeErr != nil {
			return encodeErr
		}
		for _, response := range packets {
			if sendErr := handler.Sender.Both(context.Background(), session, response); sendErr != nil {
				return sendErr
			}
		}
	}
	return nil
}

// completionPackets builds the ordered client projections for a settled trade.
func completionPackets() ([2]codec.Packet, error) {
	completed, err := outcompleted.Encode()
	if err != nil {
		return [2]codec.Packet{}, err
	}
	refresh, err := outrefresh.Encode()
	if err != nil {
		return [2]codec.Packet{}, err
	}
	return [2]codec.Packet{completed, refresh}, nil
}
