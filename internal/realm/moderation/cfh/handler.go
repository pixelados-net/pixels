// Package cfh adapts user call-for-help packets to moderation workflows.
package cfh

import (
	"context"
	"strconv"
	"time"

	moderationcore "github.com/niflaot/pixels/internal/realm/moderation/core"
	moderationrecord "github.com/niflaot/pixels/internal/realm/moderation/record"
	moderationruntime "github.com/niflaot/pixels/internal/realm/moderation/runtime"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	incfhdelete "github.com/niflaot/pixels/networking/inbound/moderation/cfh/deletepending"
	incfhpending "github.com/niflaot/pixels/networking/inbound/moderation/cfh/pending"
	inreport "github.com/niflaot/pixels/networking/inbound/moderation/cfh/report"
	inreportim "github.com/niflaot/pixels/networking/inbound/moderation/cfh/reportim"
	inselfie "github.com/niflaot/pixels/networking/inbound/moderation/cfh/selfie"
	incfhstatus "github.com/niflaot/pixels/networking/inbound/moderation/cfh/status"
	outcfhstatus "github.com/niflaot/pixels/networking/outbound/moderation/cfh/cfhsanctionstatus"
	outdisabled "github.com/niflaot/pixels/networking/outbound/moderation/cfh/disabled"
	outpending "github.com/niflaot/pixels/networking/outbound/moderation/cfh/pending"
	outdeleted "github.com/niflaot/pixels/networking/outbound/moderation/cfh/pendingdeleted"
	outresult "github.com/niflaot/pixels/networking/outbound/moderation/cfh/result"
	outcfhsanction "github.com/niflaot/pixels/networking/outbound/moderation/cfh/sanction"
	outtopics "github.com/niflaot/pixels/networking/outbound/moderation/cfh/topics"
)

// Handler adapts call-for-help protocol requests.
type Handler struct {
	// Context provides shared live moderation dependencies.
	*moderationruntime.Context
}

// Bootstrap sends call-for-help availability, topics, and sanction state.
func Bootstrap(ctx context.Context, runtime *moderationruntime.Context, playerID int64) error {
	handler := Handler{Context: runtime}
	if !handler.Moderation.Enabled() {
		packet, _ := outdisabled.Encode()
		return handler.SendTo(ctx, playerID, packet)
	}
	topics, err := handler.topicsPacket(handler.Moderation.Topics())
	if err != nil {
		return err
	}
	if err = handler.SendTo(ctx, playerID, topics); err != nil {
		return err
	}
	return RefreshStatus(ctx, runtime, playerID)
}

// RefreshStatus pushes current sanction state to one online player.
func RefreshStatus(ctx context.Context, runtime *moderationruntime.Context, playerID int64) error {
	handler := Handler{Context: runtime}
	packet, err := handler.SanctionStatus(ctx, playerID)
	if err != nil {
		return err
	}
	if err = handler.SendTo(ctx, playerID, packet); err != nil {
		return err
	}
	compatibility, _ := outcfhstatus.Encode()
	if err = handler.SendTo(ctx, playerID, compatibility); err != nil {
		return err
	}
	sanction, _ := outcfhsanction.Encode(0, 0)
	return handler.SendTo(ctx, playerID, sanction)
}

// Register installs user-facing call-for-help adapters.
func Register(registry *netconn.HandlerRegistry, runtime *moderationruntime.Context) {
	handler := Handler{Context: runtime}
	_ = registry.Register(inreport.Header, handler.callForHelp(false))
	_ = registry.Register(inreportim.Header, handler.callForHelp(true))
	_ = registry.Register(inselfie.Header, handler.selfieStub)
	_ = registry.Register(incfhpending.Header, handler.pending)
	_ = registry.Register(incfhdelete.Header, handler.deletePending)
	_ = registry.Register(incfhstatus.Header, handler.cfhStatus)
}

// callForHelp handles room or messenger evidence reports.
func (handler Handler) callForHelp(fromIM bool) netconn.Handler {
	return func(connection netconn.Context, packet codec.Packet) error {
		actorID, err := handler.Actor(connection)
		if err != nil {
			return err
		}
		params := moderationrecord.ReportParams{ReporterPlayerID: actorID, Kind: "cfh"}
		if fromIM {
			payload, decodeErr := inreportim.Decode(packet)
			if decodeErr != nil {
				return decodeErr
			}
			target := int64(payload.ReportedPlayerID)
			params.ReportedPlayerID, params.TopicID, params.Message = &target, int64(payload.TopicID), payload.Message
			for _, entry := range payload.Entries {
				params.Chatlog = append(params.Chatlog, moderationrecord.ChatEntry{PatternID: entry.Pattern, Message: entry.Message})
			}
		} else {
			payload, decodeErr := inreport.Decode(packet)
			if decodeErr != nil {
				return decodeErr
			}
			target, roomID := int64(payload.ReportedPlayerID), int64(payload.RoomID)
			params.ReportedPlayerID, params.RoomID, params.TopicID, params.Message = &target, &roomID, int64(payload.TopicID), payload.Message
			for _, entry := range payload.Entries {
				params.Chatlog = append(params.Chatlog, moderationrecord.ChatEntry{PatternID: entry.Pattern, Message: entry.Message})
			}
		}
		result, reportErr := handler.Moderation.Report(context.Background(), params)
		code, message := int32(1), handler.Text("moderation.report.received")
		if reportErr != nil {
			code, message = 2, handler.reportError(reportErr)
		} else if result.ReplyKey != "" {
			message = handler.Text(result.ReplyKey)
		} else if result.Ignored {
			message = handler.Text("moderation.report.ignored")
		}
		response, _ := outresult.Encode(code, message)
		return connection.Send(context.Background(), response)
	}
}

// selfieStub accepts the deferred packet and responds explicitly.
func (handler Handler) selfieStub(connection netconn.Context, packet codec.Packet) error {
	if _, err := inselfie.Decode(packet); err != nil {
		return err
	}
	response, _ := outresult.Encode(2, handler.Text("moderation.report.photo_unavailable"))
	return connection.Send(context.Background(), response)
}

// pending sends one reporter's unresolved calls.
func (handler Handler) pending(connection netconn.Context, packet codec.Packet) error {
	if err := incfhpending.Decode(packet); err != nil {
		return err
	}
	actorID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	issues, err := handler.Moderation.Store().Pending(context.Background(), actorID)
	if err != nil {
		return err
	}
	calls := make([]outpending.Call, len(issues))
	for index, issue := range issues {
		calls[index] = outpending.Call{ID: strconv.FormatInt(issue.ID, 10), Timestamp: issue.CreatedAt.Format(time.RFC3339), Message: issue.Message}
	}
	response, err := outpending.Encode(calls)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// deletePending removes only the actor's unresolved calls.
func (handler Handler) deletePending(connection netconn.Context, packet codec.Packet) error {
	if err := incfhdelete.Decode(packet); err != nil {
		return err
	}
	actorID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	if _, err = handler.Moderation.Store().DeletePending(context.Background(), actorID); err != nil {
		return err
	}
	response, _ := outdeleted.Encode()
	return connection.Send(context.Background(), response)
}

// cfhStatus projects availability, topics, and sanction status.
func (handler Handler) cfhStatus(connection netconn.Context, packet codec.Packet) error {
	if _, err := incfhstatus.Decode(packet); err != nil {
		return err
	}
	if !handler.Moderation.Enabled() {
		response, _ := outdisabled.Encode()
		return connection.Send(context.Background(), response)
	}
	topics, err := handler.topicsPacket(handler.Moderation.Topics())
	if err != nil {
		return err
	}
	if err = connection.Send(context.Background(), topics); err != nil {
		return err
	}
	actorID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	status, err := handler.SanctionStatus(context.Background(), actorID)
	if err != nil {
		return err
	}
	if err = connection.Send(context.Background(), status); err != nil {
		return err
	}
	compatibility, _ := outcfhstatus.Encode()
	if err = connection.Send(context.Background(), compatibility); err != nil {
		return err
	}
	sanction, _ := outcfhsanction.Encode(0, 0)
	return connection.Send(context.Background(), sanction)
}

// topicsPacket groups and localizes cached topics by category.
func (handler Handler) topicsPacket(items []moderationrecord.Topic) (codec.Packet, error) {
	categories := make([]outtopics.Category, 0)
	indexes := make(map[string]int)
	for _, item := range items {
		index, found := indexes[item.Category]
		if !found {
			index = len(categories)
			indexes[item.Category] = index
			categories = append(categories, outtopics.Category{Name: item.Category})
		}
		categories[index].Topics = append(categories[index].Topics, outtopics.Topic{Name: handler.Text(item.NameKey), ID: int32(item.ID), Consequence: item.Action})
	}
	return outtopics.Encode(categories)
}

// reportError maps expected report failures to localized text.
func (handler Handler) reportError(err error) string {
	switch err {
	case moderationcore.ErrDisabled:
		return handler.Text("moderation.report.disabled")
	case moderationcore.ErrThrottled:
		return handler.Text("moderation.report.throttled")
	default:
		return handler.Text("moderation.report.failed")
	}
}
