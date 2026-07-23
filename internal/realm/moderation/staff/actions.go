package staff

import (
	"context"
	"time"

	sanctionrecord "github.com/niflaot/pixels/internal/realm/sanction/record"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inalert "github.com/niflaot/pixels/networking/inbound/moderation/staff/sanctionalert"
	inban "github.com/niflaot/pixels/networking/inbound/moderation/staff/sanctionban"
	inkick "github.com/niflaot/pixels/networking/inbound/moderation/staff/sanctionkick"
	inmute "github.com/niflaot/pixels/networking/inbound/moderation/staff/sanctionmute"
	intrade "github.com/niflaot/pixels/networking/inbound/moderation/staff/sanctiontradelock"
	outaction "github.com/niflaot/pixels/networking/outbound/moderation/staff/actionresult"
)

const (
	// muteDurationHours is the duration of Nitro's fixed Mute 1h action.
	muteDurationHours int32 = 1
	// permanentTradeLockMinutes is Nitro's permanent trade-lock sentinel.
	permanentTradeLockMinutes int32 = 100 * 365 * 24 * 60
)

// sanctionRequest applies one decoded staff punishment through the common engine.
func (handler Handler) sanctionRequest(connection netconn.Context, targetID int64, kind sanctionrecord.Kind, reason string, hours int32, topicID int32, issueID int32) error {
	actorID, err := handler.Actor(connection)
	if err != nil {
		return err
	}
	if hours < 0 || hours > sanctionrecord.MaxDurationHours {
		response, _ := outaction.Encode(int32(targetID), false)
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
	response, _ := outaction.Encode(int32(targetID), err == nil)
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

// sanctionMute applies Nitro's fixed one-hour global mute.
func (handler Handler) sanctionMute(connection netconn.Context, packet codec.Packet) error {
	payload, err := inmute.Decode(packet)
	if err != nil {
		return err
	}
	return handler.sanctionRequest(connection, int64(payload.PlayerID), sanctionrecord.KindMute, payload.Message, muteDurationHours, payload.TopicID, payload.IssueID)
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
	hours, valid := banDurationHours(payload.ActionIndex)
	if !valid {
		return handler.rejectAction(connection, payload.PlayerID)
	}
	return handler.sanctionRequest(connection, int64(payload.PlayerID), sanctionrecord.KindBan, payload.Message, hours, payload.TopicID, payload.IssueID)
}

// sanctionTradeLock applies a direct-trade lock.
func (handler Handler) sanctionTradeLock(connection netconn.Context, packet codec.Packet) error {
	payload, err := intrade.Decode(packet)
	if err != nil {
		return err
	}
	hours, valid := tradeLockDurationHours(payload.Minutes)
	if !valid {
		return handler.rejectAction(connection, payload.PlayerID)
	}
	return handler.sanctionRequest(connection, int64(payload.PlayerID), sanctionrecord.KindTradeLock, payload.Message, hours, payload.TopicID, payload.IssueID)
}

// banDurationHours maps the actual Nitro selector index to a finite or permanent ban.
func banDurationHours(actionIndex int32) (int32, bool) {
	switch actionIndex {
	case 2:
		return 18, true
	case 3:
		return 7 * 24, true
	case 4, 5:
		return 30 * 24, true
	case 6, 7:
		return 0, true
	default:
		return 0, false
	}
}

// tradeLockDurationHours converts Nitro's minute value and recognizes its permanent sentinel.
func tradeLockDurationHours(minutes int32) (int32, bool) {
	if minutes >= permanentTradeLockMinutes {
		return 0, true
	}
	if minutes <= 0 {
		return 0, false
	}
	hours := (minutes + 59) / 60
	return hours, hours <= sanctionrecord.MaxDurationHours
}

// rejectAction sends Nitro's uniform failed-action response.
func (handler Handler) rejectAction(connection netconn.Context, targetID int32) error {
	response, _ := outaction.Encode(targetID, false)
	return connection.Send(context.Background(), response)
}
