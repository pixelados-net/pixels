package move

import (
	"context"
	"errors"

	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	roomfurniture "github.com/niflaot/pixels/internal/realm/room/furniture"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	outbubble "github.com/niflaot/pixels/networking/outbound/session/bubblealert"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/zap"
)

const (
	// bubbleKeyFurniturePlacementError matches Arcturus's BubbleAlertKeys.FURNITURE_PLACEMENT_ERROR.
	bubbleKeyFurniturePlacementError = "furni_placement_error"
)

// handleSoftError logs a rejected move attempt with context and sends a bubble alert when possible.
func (handler Handler) handleSoftError(ctx context.Context, cmd Command, err error) error {
	key, soft := bubbleErrorKey(err)
	if !soft {
		return err
	}

	if handler.Log != nil {
		handler.Log.Warn("furniture move rejected",
			zap.Int64("item_id", cmd.ItemID), zap.Int("x", cmd.X), zap.Int("y", cmd.Y), zap.Int("rotation", cmd.Rotation),
			zap.Error(err),
		)
	}
	if key == "" {
		return nil
	}

	return handler.sendBubbleAlert(ctx, cmd.Handler, key)
}

// sendBubbleAlert notifies the actor of a rejected furniture move.
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

// bubbleErrorKey reports whether an error is a soft gameplay miss and its bubble alert key, if any.
func bubbleErrorKey(err error) (string, bool) {
	switch {
	case errors.Is(err, roomlive.ErrInvalidPlacement),
		errors.Is(err, roomfurniture.ErrInvalidTarget),
		errors.Is(err, furnitureservice.ErrInvalidPlacement),
		errors.Is(err, furnitureservice.ErrInvalidItemID),
		errors.Is(err, furnitureservice.ErrInvalidRoomID),
		errors.Is(err, furnitureservice.ErrInvalidPlayerID):
		return "session.bubble.furniture.invalid_move", true
	case errors.Is(err, roomlive.ErrTileOccupied):
		return "session.bubble.furniture.tile_has_units", true
	case errors.Is(err, roomlive.ErrCannotStack):
		return "session.bubble.furniture.cant_stack", true
	case errors.Is(err, furnitureservice.ErrNotItemOwner):
		return "session.bubble.furniture.no_rights", true
	case errors.Is(err, roomlive.ErrNoFurnitureRights):
		return "session.bubble.furniture.no_rights", true
	case errors.Is(err, furnitureservice.ErrItemNotInInventory):
		return "session.bubble.furniture.item_not_in_inventory", true
	case errors.Is(err, furnitureservice.ErrItemNotFound),
		errors.Is(err, furnitureservice.ErrItemNotPlaced),
		errors.Is(err, roomfurniture.ErrDefinitionNotFound):
		return "session.bubble.furniture.item_not_found", true
	case errors.Is(err, roomlive.ErrWorldNotLoaded):
		return "", true
	default:
		return "", false
	}
}
