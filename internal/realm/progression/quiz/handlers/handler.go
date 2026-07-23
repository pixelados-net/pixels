// Package handlers adapts Nitro safety quiz requests.
package handlers

import (
	"context"

	progressionquiz "github.com/niflaot/pixels/internal/realm/progression/quiz"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inanswers "github.com/niflaot/pixels/networking/inbound/progression/quiz/answers"
	inquestions "github.com/niflaot/pixels/networking/inbound/progression/quiz/questions"
	outquestions "github.com/niflaot/pixels/networking/outbound/progression/quiz/questions"
	outresults "github.com/niflaot/pixels/networking/outbound/progression/quiz/results"
)

// Handler owns safety quiz requests.
type Handler struct {
	// Service owns quiz evaluation.
	Service *progressionquiz.Service
	// Bindings resolves authenticated players.
	Bindings *binding.Registry
}

// Register installs quiz question and answer handlers.
func Register(registry *netconn.HandlerRegistry, handler Handler) {
	if registry == nil {
		return
	}
	_ = registry.Register(inquestions.Header, handler.questions)
	_ = registry.Register(inanswers.Header, handler.answers)
}

// questions sends ordered client-known question identifiers.
func (handler Handler) questions(connection netconn.Context, packet codec.Packet) error {
	code, err := inquestions.Decode(packet)
	if err != nil {
		return err
	}
	if handler.Service == nil {
		return nil
	}
	ids, err := handler.Service.Questions(code)
	if err == progressionrecord.ErrNotFound {
		ids, err = nil, nil
	}
	if err != nil {
		return err
	}
	response, err := outquestions.Encode(code, ids)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// answers evaluates one bounded ordered quiz submission.
func (handler Handler) answers(connection netconn.Context, packet codec.Packet) error {
	request, err := inanswers.Decode(packet)
	if err != nil {
		return err
	}
	playerID, found := handler.playerID(connection)
	if !found || handler.Service == nil {
		return nil
	}
	failed, err := handler.Service.Submit(context.Background(), playerID, request.Code, request.AnswerIDs)
	if err == progressionrecord.ErrInvalid {
		return nil
	}
	if err != nil {
		return err
	}
	response, err := outresults.Encode(request.Code, failed)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// playerID resolves one authenticated connection binding.
func (handler Handler) playerID(connection netconn.Context) (int64, bool) {
	if handler.Bindings == nil {
		return 0, false
	}
	current, found := handler.Bindings.FindByConnection(binding.ConnectionKey{Kind: connection.ConnectionKind, ID: connection.ConnectionID})
	return current.PlayerID, found
}
