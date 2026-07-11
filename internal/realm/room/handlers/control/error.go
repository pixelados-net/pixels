// Package control converts expected room-control denials into localized client feedback.
package control

import (
	"context"
	"errors"

	roomcommand "github.com/niflaot/pixels/internal/realm/room/commands/control"
	roommoderation "github.com/niflaot/pixels/internal/realm/room/moderation"
	roomrights "github.com/niflaot/pixels/internal/realm/room/rights"
	roomsettings "github.com/niflaot/pixels/internal/realm/room/settings"
	roomwordfilter "github.com/niflaot/pixels/internal/realm/room/wordfilter"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outbubble "github.com/niflaot/pixels/networking/outbound/session/bubblealert"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/zap"
)

const (
	// bubbleKey identifies room control bubble alerts.
	bubbleKey = "room_control_error"
)

// Wrap converts known room-control errors into soft localized responses.
func Wrap(next netconn.Handler, translations i18n.Translator, log *zap.Logger) netconn.Handler {
	return func(connection netconn.Context, packet codec.Packet) error {
		err := next(connection, packet)
		key, soft := translationKey(err)
		if !soft {
			return err
		}
		if log != nil {
			log.Warn("room control action rejected", zap.Error(err), zap.Uint16("packet_header", packet.Header))
		}
		message := string(key)
		if translations != nil {
			message = translations.Default(key)
		}
		response, encodeErr := outbubble.Encode(bubbleKey, message, outbubble.WithDisplayBubble())
		if encodeErr != nil {
			return encodeErr
		}

		return connection.Send(context.Background(), response)
	}
}

// translationKey maps expected domain rejections to localized messages.
func translationKey(err error) (i18n.Key, bool) {
	switch {
	case errors.Is(err, roomrights.ErrAccessDenied), errors.Is(err, roommoderation.ErrAccessDenied), errors.Is(err, roomsettings.ErrAccessDenied):
		return "session.bubble.room.control.denied", true
	case errors.Is(err, roomsettings.ErrClubRequired):
		return "session.bubble.room.settings.club_required", true
	case errors.Is(err, roommoderation.ErrTargetProtected):
		return "session.bubble.room.control.protected", true
	case errors.Is(err, roommoderation.ErrTargetOwner), errors.Is(err, roomrights.ErrOwnerTarget):
		return "session.bubble.room.control.owner", true
	case errors.Is(err, roommoderation.ErrSelfTarget):
		return "session.bubble.room.control.self", true
	case errors.Is(err, roommoderation.ErrInvalidMuteDuration), errors.Is(err, roommoderation.ErrInvalidBanDuration), errors.Is(err, roomrights.ErrInvalidIdentity), errors.Is(err, roomcommand.ErrTargetNotInRoom), errors.Is(err, roomwordfilter.ErrInvalidWord):
		return "session.bubble.room.control.invalid", true
	default:
		return "", false
	}
}
