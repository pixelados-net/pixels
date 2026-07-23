package wired

import (
	"context"

	netconn "github.com/niflaot/pixels/networking/connection"
	failedpacket "github.com/niflaot/pixels/networking/outbound/furniture/wired/save/failed"
	"github.com/niflaot/pixels/pkg/i18n"
)

// sendFailure resolves and sends client-safe editor feedback.
func (handler Handler) sendFailure(ctx context.Context, connection netconn.Context, key string) error {
	message := key
	if handler.Translations != nil {
		message = handler.Translations.Default(i18n.Key(key))
	}
	packet, err := failedpacket.Encode(message)
	if err != nil {
		return err
	}
	return connection.Send(ctx, packet)
}
