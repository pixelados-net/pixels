package create

import (
	"context"
	"errors"

	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	netconn "github.com/niflaot/pixels/networking/connection"
	outcancreate "github.com/niflaot/pixels/networking/outbound/navigator/create/cancreate"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/zap"
)

// handleCreateError converts expected validation failures into client feedback.
func (handler Handler) handleCreateError(ctx context.Context, input Command, playerID int64, err error) error {
	key, soft := createErrorKey(err)
	if !soft {
		return err
	}
	if handler.Log != nil {
		handler.Log.Warn("room creation rejected",
			zap.Int64("player_id", playerID), zap.String("model_name", input.ModelName),
			zap.Int32("category_id", input.CategoryID), zap.Error(err),
		)
	}
	message := string(key)
	if handler.Translations != nil {
		message = handler.Translations.Default(key)
	}
	packet, encodeErr := outalert.Encode(message)
	if encodeErr != nil {
		return encodeErr
	}

	return input.Handler.Send(ctx, packet)
}

// sendLimit sends Nitro's native room ownership limit response.
func (handler Handler) sendLimit(ctx context.Context, connection netconn.Context) error {
	packet, err := outcancreate.Encode(1, int32(roomservice.MaxRoomsPerPlayer))
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}

// createErrorKey maps expected creation errors to localized messages.
func createErrorKey(err error) (i18n.Key, bool) {
	switch {
	case errors.Is(err, roomservice.ErrLayoutNotAvailable):
		return "navigator.create.error.layout_unavailable", true
	case errors.Is(err, roomservice.ErrInvalidCategory):
		return "navigator.create.error.category", true
	case errors.Is(err, roomservice.ErrInvalidRoomName), errors.Is(err, roomservice.ErrProhibitedName):
		return "navigator.create.error.name", true
	case errors.Is(err, roomservice.ErrInvalidDescription), errors.Is(err, roomservice.ErrProhibitedDescription):
		return "navigator.create.error.description", true
	case errors.Is(err, roomservice.ErrInvalidMaxUsers):
		return "navigator.create.error.capacity", true
	case errors.Is(err, roomservice.ErrInvalidTradeMode):
		return "navigator.create.error.trade", true
	default:
		return "", false
	}
}
