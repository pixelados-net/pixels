package staff

import (
	"context"
	"time"

	moderationcore "github.com/niflaot/pixels/internal/realm/moderation/core"
	moderationpolicy "github.com/niflaot/pixels/internal/realm/moderation/policy"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	sanctionrecord "github.com/niflaot/pixels/internal/realm/sanction/record"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inroomchange "github.com/niflaot/pixels/networking/inbound/moderation/staff/changeroom"
	inroomalert "github.com/niflaot/pixels/networking/inbound/moderation/staff/roomalert"
	inalert "github.com/niflaot/pixels/networking/inbound/moderation/staff/sanctionalert"
	inban "github.com/niflaot/pixels/networking/inbound/moderation/staff/sanctionban"
	inkick "github.com/niflaot/pixels/networking/inbound/moderation/staff/sanctionkick"
	inmute "github.com/niflaot/pixels/networking/inbound/moderation/staff/sanctionmute"
	intrade "github.com/niflaot/pixels/networking/inbound/moderation/staff/sanctiontradelock"
	outaction "github.com/niflaot/pixels/networking/outbound/moderation/staff/actionresult"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
)

// sanctionRequest applies one decoded staff punishment through the common engine.
func (handler Handler) sanctionRequest(connection netconn.Context, targetID int64, kind sanctionrecord.Kind, reason string, hours int32, topicID int32, issueID int32) error {
	actorID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	if hours < 0 || hours > sanctionrecord.MaxDurationHours {
		response, _ := outaction.Encode(1, false)
		return connection.Send(context.Background(), response)
	}
	var expiresAt *time.Time
	if hours > 0 {
		value := time.Now().Add(time.Duration(hours) * time.Hour)
		expiresAt = &value
	}
	var issue *int64
	if issueID > 0 {
		value := int64(issueID)
		issue = &value
	}
	var topic *int64
	if topicID > 0 {
		value := int64(topicID)
		topic = &value
	}
	reason = handler.Moderation.Sanitize(reason)
	_, err = handler.Sanctions.Apply(context.Background(), sanctionrecord.ApplyParams{ReceiverPlayerID: targetID, IssuerPlayerID: &actorID, IssuerKind: "player", Kind: kind, Reason: reason, CFHTopicID: topic, IssueID: issue, Source: "modtool", ExpiresAt: expiresAt})
	response, _ := outaction.Encode(actionCode(err), err == nil)
	if sendErr := connection.Send(context.Background(), response); sendErr != nil {
		return sendErr
	}
	return nil
}

// sanctionAlert applies a warning.
func (handler Handler) sanctionAlert(connection netconn.Context, packet codec.Packet) error {
	payload, err := inalert.Decode(packet)
	if err != nil {
		return err
	}
	return handler.sanctionRequest(connection, int64(payload.PlayerID), sanctionrecord.KindWarn, payload.Message, 0, payload.TopicID, payload.IssueID)
}

// sanctionMute applies a timed or permanent global mute.
func (handler Handler) sanctionMute(connection netconn.Context, packet codec.Packet) error {
	payload, err := inmute.Decode(packet)
	if err != nil {
		return err
	}
	return handler.sanctionRequest(connection, int64(payload.PlayerID), sanctionrecord.KindMute, payload.Message, payload.Hours, 0, payload.IssueID)
}

// sanctionKick applies an instantaneous hotel kick.
func (handler Handler) sanctionKick(connection netconn.Context, packet codec.Packet) error {
	payload, err := inkick.Decode(packet)
	if err != nil {
		return err
	}
	return handler.sanctionRequest(connection, int64(payload.PlayerID), sanctionrecord.KindKick, payload.Message, 0, payload.TopicID, payload.IssueID)
}

// sanctionBan applies a global login ban.
func (handler Handler) sanctionBan(connection netconn.Context, packet codec.Packet) error {
	payload, err := inban.Decode(packet)
	if err != nil {
		return err
	}
	hours := payload.Hours
	if payload.Permanent {
		hours = 0
	}
	return handler.sanctionRequest(connection, int64(payload.PlayerID), sanctionrecord.KindBan, payload.Message, hours, payload.TopicID, payload.IssueID)
}

// sanctionTradeLock applies a direct-trade lock.
func (handler Handler) sanctionTradeLock(connection netconn.Context, packet codec.Packet) error {
	payload, err := intrade.Decode(packet)
	if err != nil {
		return err
	}
	return handler.sanctionRequest(connection, int64(payload.PlayerID), sanctionrecord.KindTradeLock, payload.Message, payload.Hours, payload.TopicID, payload.IssueID)
}

// roomAlert broadcasts a localized staff warning without creating punishment history.
func (handler Handler) roomAlert(connection netconn.Context, packet codec.Packet) error {
	payload, err := inroomalert.Decode(packet)
	if err != nil {
		return err
	}
	actorID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	allowed, err := handler.Permissions.HasPermission(context.Background(), actorID, moderationpolicy.RoomOverride)
	if err != nil {
		return err
	}
	if !allowed {
		response, _ := outaction.Encode(actionCode(moderationcore.ErrUnauthorized), false)
		return connection.Send(context.Background(), response)
	}
	room, found := handler.RoomsLive.Find(int64(payload.RoomID))
	if !found {
		response, _ := outaction.Encode(1, false)
		return connection.Send(context.Background(), response)
	}
	response, err := outalert.Encode(handler.Moderation.Sanitize(payload.Message))
	if err != nil {
		return err
	}
	for _, occupant := range room.Occupants() {
		_ = handler.SendTo(context.Background(), occupant.PlayerID, response)
	}
	result, _ := outaction.Encode(0, true)
	return connection.Send(context.Background(), result)
}

// changeRoom applies protocol staff door overrides through ordinary room persistence.
func (handler Handler) changeRoom(connection netconn.Context, packet codec.Packet) error {
	payload, err := inroomchange.Decode(packet)
	if err != nil {
		return err
	}
	actorID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	allowed, err := handler.Permissions.HasPermission(context.Background(), actorID, moderationpolicy.RoomOverride)
	if err != nil {
		return err
	}
	if !allowed {
		response, _ := outaction.Encode(actionCode(moderationcore.ErrUnauthorized), false)
		return connection.Send(context.Background(), response)
	}
	room, found, err := handler.Rooms.FindByID(context.Background(), int64(payload.RoomID))
	if err != nil {
		return err
	}
	if !found {
		response, _ := outaction.Encode(1, false)
		return connection.Send(context.Background(), response)
	}
	mode := room.DoorMode
	password := (*string)(nil)
	if payload.LockDoor != 0 {
		mode = roommodel.DoorModeOpen
		empty := ""
		password = &empty
	}
	_, err = handler.Rooms.Update(context.Background(), room.ID, room.Version.Version, roomservice.UpdateParams{DoorMode: &mode, Password: password, AllowReservedTags: true})
	response, _ := outaction.Encode(actionCode(err), err == nil)
	if sendErr := connection.Send(context.Background(), response); sendErr != nil {
		return sendErr
	}
	return nil
}

// actionCode maps common errors to one stable staff response code.
func actionCode(err error) int32 {
	if err == nil {
		return 0
	}
	return 1
}
