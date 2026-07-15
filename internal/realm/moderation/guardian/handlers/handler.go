// Package handlers adapts guardian peer-review protocol packets.
package handlers

import (
	"context"
	"strings"

	historymodel "github.com/niflaot/pixels/internal/realm/chat/history/model"
	"github.com/niflaot/pixels/internal/realm/moderation/guardian"
	moderationrecord "github.com/niflaot/pixels/internal/realm/moderation/record"
	moderationruntime "github.com/niflaot/pixels/internal/realm/moderation/runtime"
	sanctioncore "github.com/niflaot/pixels/internal/realm/sanction/core"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	increate "github.com/niflaot/pixels/networking/inbound/moderation/guardian/create"
	indecide "github.com/niflaot/pixels/networking/inbound/moderation/guardian/decide"
	indetach "github.com/niflaot/pixels/networking/inbound/moderation/guardian/detach"
	invote "github.com/niflaot/pixels/networking/inbound/moderation/guardian/vote"
	outdetached "github.com/niflaot/pixels/networking/outbound/moderation/guardian/detached"
	outoffered "github.com/niflaot/pixels/networking/outbound/moderation/guardian/offered"
	outresults "github.com/niflaot/pixels/networking/outbound/moderation/guardian/results"
	outstarted "github.com/niflaot/pixels/networking/outbound/moderation/guardian/started"
)

// Disconnected detaches one reviewer and finishes a terminal ticket.
func Disconnected(ctx context.Context, runtime *moderationruntime.Context, playerID int64) error {
	handler := Handler{Context: runtime}
	ticket, found := runtime.Guardians.Detach(playerID)
	if !found {
		return nil
	}
	packet, _ := outdetached.Encode()
	for reviewerID := range ticket.Reviewers {
		_ = runtime.SendTo(ctx, reviewerID, packet)
	}
	if ticket.State == guardian.StateClosed {
		return handler.Finish(ctx, ticket)
	}
	return nil
}

// Handler adapts guardian protocol requests.
type Handler struct {
	// Context provides shared live moderation dependencies.
	*moderationruntime.Context
}

// Register installs peer-review packet adapters.
func Register(registry *netconn.HandlerRegistry, runtime *moderationruntime.Context) {
	handler := Handler{Context: runtime}
	_ = registry.Register(increate.Header, handler.guardianCreate)
	_ = registry.Register(indecide.Header, handler.guardianDecide)
	_ = registry.Register(invote.Header, handler.guardianVote)
	_ = registry.Register(indetach.Header, handler.guardianDetach)
}

// guardianCreate freezes current reported-player room history and offers reviewers.
func (handler Handler) guardianCreate(connection netconn.Context, packet codec.Packet) error {
	payload, err := increate.Decode(packet)
	if err != nil {
		return err
	}
	reporterID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	roomID, playerID := int64(payload.RoomID), int64(payload.ReportedPlayerID)
	entries, err := handler.History.History(context.Background(), historymodel.Query{RoomID: &roomID, PlayerID: &playerID, Limit: 50})
	if err != nil {
		return err
	}
	chatlog := make([]moderationrecord.ChatEntry, len(entries))
	for index, entry := range entries {
		id := entry.PlayerID
		chatlog[index] = moderationrecord.ChatEntry{PlayerID: &id, PatternID: entry.Kind, Message: entry.Message, CreatedAt: entry.CreatedAt}
	}
	ticket, err := handler.Guardians.Create(context.Background(), reporterID, playerID, chatlog)
	if err != nil {
		return err
	}
	response, _ := outoffered.Encode(int32(ticket.ID))
	for reviewerID := range ticket.Reviewers {
		_ = handler.SendTo(context.Background(), reviewerID, response)
	}
	return nil
}

// guardianDecide records an offer decision and starts accepted reviewers.
func (handler Handler) guardianDecide(connection netconn.Context, packet codec.Packet) error {
	payload, err := indecide.Decode(packet)
	if err != nil {
		return err
	}
	actorID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	ticket, err := handler.Guardians.Decide(actorID, payload.Accepted)
	if err != nil {
		return err
	}
	if ticket.State != guardian.StateVoting {
		if ticket.State == guardian.StateClosed {
			response, _ := outdetached.Encode()
			return handler.SendTo(context.Background(), ticket.ReporterPlayerID, response)
		}
		return nil
	}
	response, _ := outstarted.Encode(int32(ticket.ID), guardianTranscript(ticket.Chatlog))
	for id, reviewer := range ticket.Reviewers {
		if reviewer.Accepted {
			_ = handler.SendTo(context.Background(), id, response)
		}
	}
	return nil
}

// guardianVote records and finalizes one reviewer vote.
func (handler Handler) guardianVote(connection netconn.Context, packet codec.Packet) error {
	payload, err := invote.Decode(packet)
	if err != nil {
		return err
	}
	actorID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	ticket, complete, err := handler.Guardians.Vote(actorID, guardian.Verdict(payload.Vote))
	if err != nil || !complete {
		return err
	}
	return handler.Finish(context.Background(), ticket)
}

// guardianDetach removes one active reviewer.
func (handler Handler) guardianDetach(connection netconn.Context, packet codec.Packet) error {
	if err := indetach.Decode(packet); err != nil {
		return err
	}
	actorID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	ticket, found := handler.Guardians.Detach(actorID)
	if found {
		response, _ := outdetached.Encode()
		if err = connection.Send(context.Background(), response); err != nil {
			return err
		}
		if ticket.State == guardian.StateClosed {
			return handler.Finish(context.Background(), ticket)
		}
	}
	return nil
}

// Finish publishes results and escalates only on strict actionable majority.
func (handler Handler) Finish(ctx context.Context, ticket guardian.Ticket) error {
	response, _ := outresults.Encode(int32(ticket.ID), int32(ticket.Result), int32(len(ticket.Reviewers)), int32(len(ticket.Reviewers)))
	for id := range ticket.Reviewers {
		_ = handler.SendTo(ctx, id, response)
	}
	_ = handler.SendTo(ctx, ticket.ReporterPlayerID, response)
	if ticket.Result == guardian.VerdictBad || ticket.Result == guardian.VerdictHorrible {
		_, err := handler.Sanctions.EscalateFor(ctx, sanctioncore.EscalateParams{ReceiverPlayerID: ticket.ReportedPlayerID, Reason: "guardian chat review", Source: "cfh_auto"})
		return err
	}
	if ticket.Result == guardian.VerdictMixed {
		topicID := int64(1)
		_, err := handler.Moderation.Report(ctx, moderationrecord.ReportParams{ReporterPlayerID: ticket.ReporterPlayerID, ReportedPlayerID: &ticket.ReportedPlayerID, TopicID: topicID, Kind: "guardian", Message: "mixed guardian review", Chatlog: ticket.Chatlog})
		return err
	}
	return nil
}

// guardianTranscript serializes anonymized evidence for Nitro's compact field.
func guardianTranscript(entries []moderationrecord.ChatEntry) string {
	parts := make([]string, 0, len(entries))
	for _, entry := range entries {
		parts = append(parts, entry.PatternID+": "+entry.Message)
	}
	return strings.Join(parts, "\n")
}
