package pickup

import (
	"context"
	"errors"

	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	outbubble "github.com/niflaot/pixels/networking/outbound/session/bubblealert"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/zap"
)

// handleSoftError logs an expected pickup rejection and sends localized feedback.
func (handler Handler) handleSoftError(ctx context.Context, cmd Command, err error) error {
	key, soft := bubbleErrorKey(err)
	if !soft {
		return err
	}
	if handler.Log != nil {
		handler.Log.Warn("furniture pickup rejected", zap.Int64("item_id", cmd.ItemID), zap.Error(err))
	}
	if key == "" {
		return nil
	}

	return handler.sendBubbleAlert(ctx, cmd.Handler, key)
}

// sendBubbleAlert notifies the actor of a rejected furniture pickup.
func (handler Handler) sendBubbleAlert(ctx context.Context, connection netconn.Context, key string) error {
	message := string(key)
	if handler.Translations != nil {
		message = handler.Translations.Default(i18n.Key(key))
	}
	packet, err := outbubble.Encode(bubbleKeyFurniturePlacementError, message, outbubble.WithDisplayBubble())
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}

// bubbleErrorKey reports whether an error is a soft gameplay miss and its bubble alert key.
func bubbleErrorKey(err error) (string, bool) {
	switch {
	case errors.Is(err, furnitureservice.ErrNotItemOwner), errors.Is(err, roomlive.ErrNoFurnitureRights):
		return "session.bubble.furniture.no_rights", true
	case errors.Is(err, furnitureservice.ErrItemNotFound), errors.Is(err, furnitureservice.ErrItemNotPlaced):
		return "session.bubble.furniture.item_not_found", true
	case errors.Is(err, furnitureservice.ErrInvalidItemID), errors.Is(err, furnitureservice.ErrInvalidPlayerID):
		return "session.bubble.furniture.invalid_move", true
	default:
		return "", false
	}
}
