// Package handlers adapts Nitro room word-poll packets.
package handlers

import (
	"context"
	"errors"

	progressionpoll "github.com/niflaot/pixels/internal/realm/progression/poll"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inanswer "github.com/niflaot/pixels/networking/inbound/progression/poll/answer"
	inreject "github.com/niflaot/pixels/networking/inbound/progression/poll/reject"
	instart "github.com/niflaot/pixels/networking/inbound/progression/poll/start"
	outerror "github.com/niflaot/pixels/networking/outbound/progression/poll/error"
)

// Handler owns room word-poll client requests.
type Handler struct {
	// Service owns ephemeral poll state.
	Service *progressionpoll.Service
	// Bindings resolves authenticated players.
	Bindings *binding.Registry
}

// Register installs every room word-poll adapter.
func Register(registry *netconn.HandlerRegistry, handler Handler) {
	if registry == nil {
		return
	}
	_ = registry.Register(instart.Header, handler.handle)
	_ = registry.Register(inanswer.Header, handler.handle)
	_ = registry.Register(inreject.Header, handler.handle)
}

// handle dispatches one bounded poll request.
func (handler Handler) handle(connection netconn.Context, packet codec.Packet) error {
	playerID, found := handler.playerID(connection)
	if !found || handler.Service == nil {
		return nil
	}
	ctx := context.Background()
	var err error
	switch packet.Header {
	case instart.Header:
		var pollID int32
		pollID, err = instart.Decode(packet)
		if err == nil {
			var response codec.Packet
			var current bool
			response, current, err = handler.Service.DatabaseContents(ctx, playerID, pollID)
			if err == nil && current {
				return connection.Send(ctx, response)
			}
			if err == nil {
				response, current, err = handler.Service.Current(playerID, pollID)
			}
			if err == nil && current {
				return connection.Send(ctx, response)
			}
		}
	case inanswer.Header:
		var request inanswer.Request
		request, err = inanswer.Decode(packet)
		if err == nil {
			var durable bool
			durable, err = handler.Service.HasDatabasePoll(ctx, request.PollID)
			if err == nil && durable {
				err = handler.Service.AnswerDatabase(ctx, playerID, request.PollID, request.QuestionID, request.Values)
			} else if err == nil {
				err = handler.Service.Answer(ctx, playerID, request.PollID, request.QuestionID, request.Values)
			}
		}
	case inreject.Header:
		var pollID int32
		pollID, err = inreject.Decode(packet)
		if err == nil {
			var durable bool
			durable, err = handler.Service.HasDatabasePoll(ctx, pollID)
			if err == nil && durable {
				err = handler.Service.RejectDatabase(ctx, playerID, pollID)
			} else if err == nil {
				err = handler.Service.Reject(playerID, pollID)
			}
		}
	default:
		return codec.ErrUnexpectedHeader
	}
	if errors.Is(err, progressionpoll.ErrUnavailable) || errors.Is(err, progressionpoll.ErrForbidden) || errors.Is(err, progressionpoll.ErrInvalidAnswer) {
		response, encodeErr := outerror.Encode()
		if encodeErr != nil {
			return encodeErr
		}
		return connection.Send(ctx, response)
	}
	return err
}

// playerID resolves the authenticated request source.
func (handler Handler) playerID(connection netconn.Context) (int64, bool) {
	if handler.Bindings == nil {
		return 0, false
	}
	value, found := handler.Bindings.FindByConnection(binding.ConnectionKey{Kind: connection.ConnectionKind, ID: connection.ConnectionID})
	return value.PlayerID, found
}
