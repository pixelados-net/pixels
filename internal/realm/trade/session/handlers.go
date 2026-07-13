// Package session adapts direct-trade lifecycle packets.
package session

import (
	"context"
	"errors"
	tradecore "github.com/niflaot/pixels/internal/realm/trade/core"
	traderuntime "github.com/niflaot/pixels/internal/realm/trade/runtime"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	incancel "github.com/niflaot/pixels/networking/inbound/trade/cancel"
	inclose "github.com/niflaot/pixels/networking/inbound/trade/close"
	instart "github.com/niflaot/pixels/networking/inbound/trade/start"
	outclosed "github.com/niflaot/pixels/networking/outbound/trade/closed"
	outopen "github.com/niflaot/pixels/networking/outbound/trade/open"
	outopenfailed "github.com/niflaot/pixels/networking/outbound/trade/openfailed"
	outothernotallowed "github.com/niflaot/pixels/networking/outbound/trade/othernotallowed"
	outyounotallowed "github.com/niflaot/pixels/networking/outbound/trade/younotallowed"
)

// Handler owns direct-trade lifecycle adapters.
type Handler struct {
	// Service executes direct-trade lifecycle behavior.
	Service *tradecore.Service
	// Sender projects packets to both participants.
	Sender *traderuntime.Sender
}

// Register installs direct-trade lifecycle handlers.
func Register(registry *netconn.HandlerRegistry, handler Handler) {
	_ = registry.Register(instart.Header, handler.start)
	_ = registry.Register(inclose.Header, handler.close)
	_ = registry.Register(incancel.Header, handler.cancel)
}

// start opens a direct trade.
func (handler Handler) start(connection netconn.Context, packet codec.Packet) error {
	target, err := instart.Decode(packet)
	if err != nil {
		return err
	}
	playerID, err := handler.Sender.PlayerID(connection)
	if err != nil {
		return err
	}
	session, serviceErr := handler.Service.Start(context.Background(), playerID, int64(target), connection.RemoteAddr)
	if serviceErr != nil {
		response, encodeErr := startFailurePacket(serviceErr)
		if encodeErr != nil {
			return encodeErr
		}
		return connection.Send(context.Background(), response)
	}
	session.SetIP(session.Second.PlayerID, tradecore.NormalizeIP(handler.Sender.RemoteAddr(session.Second.PlayerID)))
	response, err := openPacket(session)
	if err != nil {
		return err
	}
	return handler.Sender.Both(context.Background(), session, response)
}

// startFailurePacket maps direct-trade failures to Nitro notification semantics.
func startFailurePacket(serviceErr error) (codec.Packet, error) {
	switch {
	case errors.Is(serviceErr, tradecore.ErrActorNotAllowed):
		return outyounotallowed.Encode()
	case errors.Is(serviceErr, tradecore.ErrTargetNotAllowed):
		return outothernotallowed.Encode()
	case errors.Is(serviceErr, tradecore.ErrRoomPolicy):
		return outopenfailed.Encode(6, "")
	case errors.Is(serviceErr, tradecore.ErrThrottled):
		return outopenfailed.Encode(7, "")
	case errors.Is(serviceErr, tradecore.ErrDisabled):
		return outopenfailed.Encode(1, "")
	default:
		return outopenfailed.Encode(8, "")
	}
}

// openPacket projects durable player identifiers expected by Nitro's user data manager.
func openPacket(session *traderuntime.Session) (codec.Packet, error) {
	return outopen.Encode(session.First.PlayerID, true, session.Second.PlayerID, true)
}

// close closes one active trade.
func (handler Handler) close(connection netconn.Context, packet codec.Packet) error {
	if err := inclose.Decode(packet); err != nil {
		return err
	}
	return handler.finish(connection)
}

// cancel confirms cancellation of one active trade.
func (handler Handler) cancel(connection netconn.Context, packet codec.Packet) error {
	if err := incancel.Decode(packet); err != nil {
		return err
	}
	return handler.finish(connection)
}

// finish removes one session and notifies both sides.
func (handler Handler) finish(connection netconn.Context) error {
	playerID, err := handler.Sender.PlayerID(connection)
	if err != nil {
		return err
	}
	session, found := handler.Service.Registry().Find(playerID)
	if !found {
		return nil
	}
	participant, _ := session.Participant(playerID)
	if !handler.Service.Close(playerID) {
		return nil
	}
	response, _ := outclosed.Encode(participant.PlayerID, 0)
	return handler.Sender.Both(context.Background(), session, response)
}
