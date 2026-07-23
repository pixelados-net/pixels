package send

import (
	"context"
	"math"
	"time"

	netconn "github.com/niflaot/pixels/networking/connection"
	outflood "github.com/niflaot/pixels/networking/outbound/chat/flood"
	outmute "github.com/niflaot/pixels/networking/outbound/room/moderation/muted"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	"github.com/niflaot/pixels/pkg/i18n"
)

// sendMute sends Nitro's native remaining-mute feedback.
func (service *Service) sendMute(ctx context.Context, connection netconn.Context, remaining time.Duration) error {
	seconds := int32(math.Ceil(remaining.Seconds()))
	packet, err := outmute.Encode(max(0, seconds))
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}

// sendFlood sends Nitro's native flood cooldown feedback.
func (service *Service) sendFlood(ctx context.Context, connection netconn.Context, remaining time.Duration) error {
	packet, err := outflood.Encode(int32(math.Ceil(remaining.Seconds())))
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}

// sendAlert sends localized expected validation feedback.
func (service *Service) sendAlert(ctx context.Context, connection netconn.Context, key string) error {
	message := key
	if service.translations != nil {
		message = service.translations.Default(i18n.Key(key))
	}
	packet, err := outalert.Encode(message)
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}

// sendAlertConnection sends localized feedback through a resolved connection.
func (service *Service) sendAlertConnection(ctx context.Context, connection netconn.Connection, key string) error {
	message := key
	if service.translations != nil {
		message = service.translations.Default(i18n.Key(key))
	}
	packet, err := outalert.Encode(message)
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}
