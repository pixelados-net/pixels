// Package handlers adapts guide-session protocol packets.
package handlers

import (
	"context"

	sessioncompleted "github.com/niflaot/pixels/internal/realm/moderation/events/sessioncompleted"
	moderationpolicy "github.com/niflaot/pixels/internal/realm/moderation/policy"
	moderationruntime "github.com/niflaot/pixels/internal/realm/moderation/runtime"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	incancel "github.com/niflaot/pixels/networking/inbound/moderation/guide/cancel"
	increate "github.com/niflaot/pixels/networking/inbound/moderation/guide/create"
	indecide "github.com/niflaot/pixels/networking/inbound/moderation/guide/decide"
	induty "github.com/niflaot/pixels/networking/inbound/moderation/guide/duty"
	infeedback "github.com/niflaot/pixels/networking/inbound/moderation/guide/feedback"
	ininvite "github.com/niflaot/pixels/networking/inbound/moderation/guide/invite"
	inmessage "github.com/niflaot/pixels/networking/inbound/moderation/guide/message"
	inreport "github.com/niflaot/pixels/networking/inbound/moderation/guide/report"
	instatus "github.com/niflaot/pixels/networking/inbound/moderation/guide/reportingstatus"
	inrequesterroom "github.com/niflaot/pixels/networking/inbound/moderation/guide/requesterroom"
	inresolve "github.com/niflaot/pixels/networking/inbound/moderation/guide/resolve"
	intyping "github.com/niflaot/pixels/networking/inbound/moderation/guide/typing"
	outattached "github.com/niflaot/pixels/networking/outbound/moderation/guide/attached"
	outcreation "github.com/niflaot/pixels/networking/outbound/moderation/guide/creationresult"
	outduty "github.com/niflaot/pixels/networking/outbound/moderation/guide/dutystatus"
	outended "github.com/niflaot/pixels/networking/outbound/moderation/guide/ended"
	outerror "github.com/niflaot/pixels/networking/outbound/moderation/guide/error"
	outmessage "github.com/niflaot/pixels/networking/outbound/moderation/guide/message"
	outstarted "github.com/niflaot/pixels/networking/outbound/moderation/guide/started"
	outtyping "github.com/niflaot/pixels/networking/outbound/moderation/guide/typing"
	"github.com/niflaot/pixels/pkg/bus"
)

// Handler adapts guide protocol requests.
type Handler struct {
	// Context provides shared live moderation dependencies.
	*moderationruntime.Context
}

// Register installs all guide-session packet adapters.
func Register(registry *netconn.HandlerRegistry, runtime *moderationruntime.Context) {
	handler := Handler{Context: runtime}
	_ = registry.Register(induty.Header, handler.guideDuty)
	_ = registry.Register(increate.Header, handler.guideCreate)
	_ = registry.Register(indecide.Header, handler.guideDecide)
	_ = registry.Register(inmessage.Header, handler.guideMessage)
	_ = registry.Register(intyping.Header, handler.guideTyping)
	_ = registry.Register(incancel.Header, handler.guideEnd)
	_ = registry.Register(inresolve.Header, handler.guideEnd)
	_ = registry.Register(infeedback.Header, handler.guideFeedback)
	_ = registry.Register(ininvite.Header, handler.guideInvite)
	_ = registry.Register(inrequesterroom.Header, handler.guideRequesterRoom)
	_ = registry.Register(inreport.Header, handler.guideReport)
	_ = registry.Register(instatus.Header, handler.guideStatus)
}

// guideDuty updates authorized helper pool participation.
func (handler Handler) guideDuty(connection netconn.Context, packet codec.Packet) error {
	payload, err := induty.Decode(packet)
	if err != nil {
		return err
	}
	actorID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	guideAllowed, err := handler.Permissions.HasPermission(context.Background(), actorID, moderationpolicy.GuideDuty)
	if err != nil {
		return err
	}
	guardianAllowed, err := handler.Permissions.HasPermission(context.Background(), actorID, moderationpolicy.GuardianDuty)
	if err != nil {
		return err
	}
	handler.Guides.SetDuty(actorID, guideAllowed && payload.Guide, guideAllowed && payload.Bully, guardianAllowed && payload.Guardian)
	guides, bullies, guardians := handler.Guides.DutyCount()
	active := guideAllowed && (payload.Guide || payload.Bully) || guardianAllowed && payload.Guardian
	response, _ := outduty.Encode(active, guides, bullies, guardians)
	return connection.Send(context.Background(), response)
}

// guideCreate matches a requester to the oldest idle guide.
func (handler Handler) guideCreate(connection netconn.Context, packet codec.Packet) error {
	payload, err := increate.Decode(packet)
	if err != nil {
		return err
	}
	actorID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	session, err := handler.Guides.Create(actorID, payload.Topic, payload.Description)
	if err != nil {
		response, _ := outerror.Encode(1)
		return connection.Send(context.Background(), response)
	}
	record, _, _ := handler.PlayerRecords.FindByID(context.Background(), actorID)
	attached, _ := outattached.Encode(true, int32(actorID), record.Player.Username, payload.Topic)
	_ = handler.SendTo(context.Background(), session.GuidePlayerID, attached)
	created, _ := outcreation.Encode(0)
	return connection.Send(context.Background(), created)
}

// guideDecide accepts or rematches one request.
func (handler Handler) guideDecide(connection netconn.Context, packet codec.Packet) error {
	payload, err := indecide.Decode(packet)
	if err != nil {
		return err
	}
	actorID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	session, err := handler.Guides.Decide(actorID, payload.Accepted)
	if err != nil {
		response, _ := outerror.Encode(1)
		_ = handler.SendTo(context.Background(), session.RequesterPlayerID, response)
		return connection.Send(context.Background(), response)
	}
	if !payload.Accepted {
		record, _, _ := handler.PlayerRecords.FindByID(context.Background(), session.RequesterPlayerID)
		attached, _ := outattached.Encode(true, int32(session.RequesterPlayerID), record.Player.Username, session.Topic)
		return handler.SendTo(context.Background(), session.GuidePlayerID, attached)
	}
	requester, _, _ := handler.PlayerRecords.FindByID(context.Background(), session.RequesterPlayerID)
	guideRecord, _, _ := handler.PlayerRecords.FindByID(context.Background(), session.GuidePlayerID)
	toRequester, _ := outstarted.Encode(int32(session.GuidePlayerID), guideRecord.Player.Username, guideRecord.Profile.Look, session.Topic, session.Description, session.CreatedAt.String())
	toGuide, _ := outstarted.Encode(int32(session.RequesterPlayerID), requester.Player.Username, requester.Profile.Look, session.Topic, session.Description, session.CreatedAt.String())
	_ = handler.SendTo(context.Background(), session.RequesterPlayerID, toRequester)
	return handler.SendTo(context.Background(), session.GuidePlayerID, toGuide)
}

// guideMessage filters, records, and relays one session message.
func (handler Handler) guideMessage(connection netconn.Context, packet codec.Packet) error {
	payload, err := inmessage.Decode(packet)
	if err != nil {
		return err
	}
	actorID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	session, message, err := handler.Guides.Send(actorID, payload.Message)
	if err != nil || message.Text == "" {
		return err
	}
	partner, _ := session.Partner(actorID)
	response, _ := outmessage.Encode(message.Text, int32(actorID))
	return handler.SendTo(context.Background(), partner, response)
}

// guideTyping relays transient typing state.
func (handler Handler) guideTyping(connection netconn.Context, packet codec.Packet) error {
	payload, err := intyping.Decode(packet)
	if err != nil {
		return err
	}
	actorID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	session, found := handler.Guides.SessionFor(actorID)
	if !found {
		return nil
	}
	partner, _ := session.Partner(actorID)
	response, _ := outtyping.Encode(payload.Typing)
	return handler.SendTo(context.Background(), partner, response)
}

// guideEnd closes either cancellation or successful resolution.
func (handler Handler) guideEnd(connection netconn.Context, packet codec.Packet) error {
	if packet.Header == incancel.Header {
		if err := incancel.Decode(packet); err != nil {
			return err
		}
	} else if err := inresolve.Decode(packet); err != nil {
		return err
	}
	actorID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	session, found := handler.Guides.End(actorID)
	if !found {
		return nil
	}
	response, _ := outended.Encode(0)
	_ = handler.SendTo(context.Background(), session.RequesterPlayerID, response)
	return handler.SendTo(context.Background(), session.GuidePlayerID, response)
}

// guideFeedback persists one requester's recommendation.
func (handler Handler) guideFeedback(connection netconn.Context, packet codec.Packet) error {
	payload, err := infeedback.Decode(packet)
	if err != nil {
		return err
	}
	actorID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	session, found := handler.Guides.TakeCompleted(actorID)
	if !found {
		return nil
	}
	if err = handler.Moderation.Store().InsertFeedback(context.Background(), session.GuidePlayerID, actorID, payload.Recommended); err != nil {
		return err
	}
	if handler.Events != nil {
		_ = handler.Events.Publish(context.Background(), bus.Event{Name: sessioncompleted.Name, Payload: sessioncompleted.Payload{GuideID: session.GuidePlayerID, RequesterID: actorID, Feedback: payload.Recommended}})
	}
	return nil
}
