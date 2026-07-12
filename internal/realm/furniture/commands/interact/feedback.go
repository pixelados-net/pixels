package interact

import (
	"context"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	outbubble "github.com/niflaot/pixels/networking/outbound/session/bubblealert"
	"github.com/niflaot/pixels/pkg/i18n"
)

const (
	// bubbleKeyFurniturePlacementError identifies Nitro's furniture feedback bubble.
	bubbleKeyFurniturePlacementError = "furni_placement_error"
	// noRightsTranslationKey identifies localized furniture authorization feedback.
	noRightsTranslationKey i18n.Key = "session.bubble.furniture.no_rights"
)

// sendNoRights sends localized interaction authorization feedback.
func (handler Handler) sendNoRights(ctx context.Context, connection netconn.Context) error {
	message := string(noRightsTranslationKey)
	if handler.Translations != nil {
		message = handler.Translations.Default(noRightsTranslationKey)
	}
	packet, err := outbubble.Encode(bubbleKeyFurniturePlacementError, message, outbubble.WithDisplayBubble())
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}

// resync restores runtime and clients after a concurrent durable state mutation.
func (handler Handler) resync(ctx context.Context, active *roomlive.Room, itemID int64, interactionType string) error {
	item, found, err := handler.Furniture.FindItemByID(ctx, itemID)
	if err != nil || !found || item.RoomID == nil || *item.RoomID != active.ID() {
		return err
	}
	updated, err := active.UpdateFurnitureState(itemID, item.ExtraData, interactionType == "gate")
	if err != nil {
		return err
	}

	return handler.broadcast(ctx, active, updated.ID, updated.ExtraData)
}
