// Package staff adapts global moderator-tool packets.
package staff

import (
	"context"

	moderationpolicy "github.com/niflaot/pixels/internal/realm/moderation/policy"
	moderationruntime "github.com/niflaot/pixels/internal/realm/moderation/runtime"
	staffroom "github.com/niflaot/pixels/internal/realm/moderation/staff/room"
	sanctioncore "github.com/niflaot/pixels/internal/realm/sanction/core"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inalertevent "github.com/niflaot/pixels/networking/inbound/moderation/staff/alertevent"
	incfhlog "github.com/niflaot/pixels/networking/inbound/moderation/staff/cfhchatlog"
	incloseaction "github.com/niflaot/pixels/networking/inbound/moderation/staff/closeissue"
	inclose "github.com/niflaot/pixels/networking/inbound/moderation/staff/closeissues"
	indefault "github.com/niflaot/pixels/networking/inbound/moderation/staff/defaultsanction"
	inpick "github.com/niflaot/pixels/networking/inbound/moderation/staff/pickissues"
	inprefs "github.com/niflaot/pixels/networking/inbound/moderation/staff/preferences"
	inrelease "github.com/niflaot/pixels/networking/inbound/moderation/staff/releaseissues"
	inroomlog "github.com/niflaot/pixels/networking/inbound/moderation/staff/roomchatlog"
	inroominfo "github.com/niflaot/pixels/networking/inbound/moderation/staff/roominfo"
	insanction "github.com/niflaot/pixels/networking/inbound/moderation/staff/sanction"
	inalert "github.com/niflaot/pixels/networking/inbound/moderation/staff/sanctionalert"
	inban "github.com/niflaot/pixels/networking/inbound/moderation/staff/sanctionban"
	inkick "github.com/niflaot/pixels/networking/inbound/moderation/staff/sanctionkick"
	inmute "github.com/niflaot/pixels/networking/inbound/moderation/staff/sanctionmute"
	intrade "github.com/niflaot/pixels/networking/inbound/moderation/staff/sanctiontradelock"
	inuserlog "github.com/niflaot/pixels/networking/inbound/moderation/staff/userchatlog"
	inuserinfo "github.com/niflaot/pixels/networking/inbound/moderation/staff/userinfo"
	invisits "github.com/niflaot/pixels/networking/inbound/moderation/staff/userrooms"
	outaction "github.com/niflaot/pixels/networking/outbound/moderation/staff/actionresult"
	"github.com/niflaot/pixels/networking/outbound/moderation/staff/issueinfo"
	outmessage "github.com/niflaot/pixels/networking/outbound/moderation/staff/message"
	outtool "github.com/niflaot/pixels/networking/outbound/moderation/staff/tool"
)

// Handler adapts staff protocol requests.
type Handler struct {
	// Context provides shared live moderation dependencies.
	*moderationruntime.Context
}

// Bootstrap sends the moderator tool only to authorized online staff.
func Bootstrap(ctx context.Context, runtime *moderationruntime.Context, playerID int64) error {
	handler := Handler{Context: runtime}
	if !handler.Moderation.Enabled() {
		return nil
	}
	allowed, err := handler.Permissions.HasPermission(ctx, playerID, moderationpolicy.ToolAccess)
	if err != nil || !allowed {
		return err
	}
	issues, err := handler.Moderation.Store().Issues(ctx, "", 100)
	if err != nil {
		return err
	}
	wireIssues := make([]issueinfo.Params, 0, len(issues))
	for _, issue := range issues {
		wireIssues = append(wireIssues, issueParams(issue))
	}
	presets, err := handler.Moderation.Store().Presets(ctx, false)
	if err != nil {
		return err
	}
	messages := make([]string, len(presets))
	for index, preset := range presets {
		messages[index] = handler.Text(preset.MessageKey)
	}
	chatlogs, _ := handler.Permissions.HasPermission(ctx, playerID, moderationpolicy.ChatlogRead)
	canBan, _ := handler.Permissions.HasPermission(ctx, playerID, sanctioncore.BanNode)
	canApply, _ := handler.Permissions.HasPermission(ctx, playerID, sanctioncore.ApplyNode)
	roomOverride, _ := handler.Permissions.HasPermission(ctx, playerID, moderationpolicy.RoomOverride)
	packet, err := outtool.Encode(wireIssues, messages, messages, outtool.Permissions{CFH: true, Chatlogs: chatlogs, Alert: canApply, Kick: canApply, Ban: canBan, RoomAlert: roomOverride, RoomKick: roomOverride})
	if err != nil {
		return err
	}
	return handler.SendTo(ctx, playerID, packet)
}

// Register installs global moderator tool adapters.
func Register(registry *netconn.HandlerRegistry, runtime *moderationruntime.Context) {
	handler := Handler{Context: runtime}
	_ = registry.Register(inpick.Header, handler.pickIssues)
	_ = registry.Register(inrelease.Header, handler.releaseIssues)
	_ = registry.Register(inclose.Header, handler.closeIssues)
	_ = registry.Register(incloseaction.Header, handler.closeDefault)
	_ = registry.Register(inprefs.Header, handler.preferences)
	_ = registry.Register(incfhlog.Header, handler.cfhChatlog)
	_ = registry.Register(inroomlog.Header, handler.roomChatlog)
	_ = registry.Register(inroominfo.Header, handler.roomInfo)
	_ = registry.Register(inuserlog.Header, handler.userChatlog)
	_ = registry.Register(inuserinfo.Header, handler.userInfo)
	_ = registry.Register(invisits.Header, handler.visits)
	_ = registry.Register(inalert.Header, handler.sanctionAlert)
	_ = registry.Register(inmute.Header, handler.sanctionMute)
	_ = registry.Register(inkick.Header, handler.sanctionKick)
	_ = registry.Register(inban.Header, handler.sanctionBan)
	_ = registry.Register(intrade.Header, handler.sanctionTradeLock)
	staffroom.Register(registry, runtime)
	_ = registry.Register(indefault.Header, handler.defaultSanction)
	_ = registry.Register(inalertevent.Header, handler.messageEvent)
	_ = registry.Register(insanction.Header, handler.sanctionCompatibility)
}

// defaultSanction applies the configured escalation policy without an issue close.
func (handler Handler) defaultSanction(connection netconn.Context, packet codec.Packet) error {
	payload, err := indefault.Decode(packet)
	if err != nil {
		return err
	}
	actorID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	allowed, err := handler.Permissions.HasPermission(context.Background(), actorID, sanctioncore.ApplyNode)
	if err == nil && allowed {
		_, err = handler.Sanctions.EscalateFor(context.Background(), sanctioncore.EscalateParams{ReceiverPlayerID: int64(payload.PlayerID), Reason: handler.Moderation.Sanitize(payload.Reason), Source: "escalation"})
	} else if err == nil {
		err = sanctioncore.ErrUnauthorized
	}
	response, _ := outaction.Encode(payload.PlayerID, err == nil)
	return connection.Send(context.Background(), response)
}

// messageEvent delivers Nitro's direct moderator message without creating a punishment.
func (handler Handler) messageEvent(connection netconn.Context, packet codec.Packet) error {
	payload, err := inalertevent.Decode(packet)
	if err != nil {
		return err
	}
	actorID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	allowed, err := handler.Permissions.HasPermission(context.Background(), actorID, sanctioncore.ApplyNode)
	message := handler.Moderation.Sanitize(payload.Message)
	success := err == nil && allowed && message != ""
	if success {
		_, success = handler.Players.Find(int64(payload.PlayerID))
	}
	if success {
		notice, encodeErr := outmessage.Encode(message, "")
		if encodeErr != nil {
			success = false
		} else if sendErr := handler.SendTo(context.Background(), int64(payload.PlayerID), notice); sendErr != nil {
			success = false
		}
	}
	response, _ := outaction.Encode(payload.PlayerID, success)
	return connection.Send(context.Background(), response)
}

// sanctionCompatibility applies Nitro's selected sanction-ladder level.
func (handler Handler) sanctionCompatibility(connection netconn.Context, packet codec.Packet) error {
	payload, err := insanction.Decode(packet)
	if err != nil {
		return err
	}
	entries, err := handler.Sanctions.Store().Ladder(context.Background())
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.Level == payload.SanctionID {
			return handler.sanctionRequest(connection, int64(payload.PlayerID), entry.Kind, handler.Text("moderation.sanction.default_reason"), entry.DurationHours, payload.TopicID, 0)
		}
	}
	response, _ := outaction.Encode(payload.PlayerID, false)
	return connection.Send(context.Background(), response)
}
