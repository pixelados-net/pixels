package floorplan

import (
	"context"
	"errors"
	"strings"

	domain "github.com/niflaot/pixels/internal/realm/room/control/floorplan"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	netconn "github.com/niflaot/pixels/networking/connection"
	outbubble "github.com/niflaot/pixels/networking/outbound/session/bubblealert"
	"github.com/niflaot/pixels/pkg/i18n"
)

const (
	// bubbleKey identifies Nitro's native floor plan editor error presentation.
	bubbleKey = "floorplan_editor.error"
)

// RoomFinder reads room records for floor plan commands.
type RoomFinder interface {
	// FindByID finds one active room record.
	FindByID(context.Context, int64) (roommodel.Room, bool, error)
}

// sendError sends expected floor plan failures without disconnecting the player.
func sendError(ctx context.Context, connection netconn.Context, cause error, translations i18n.Translator) error {
	keys := errorKeys(cause)
	if len(keys) == 0 {
		return cause
	}
	messages := make([]string, len(keys))
	for index, key := range keys {
		messages[index] = string(key)
		if translations != nil {
			messages[index] = translations.Default(key)
		}
	}
	packet, err := outbubble.Encode(bubbleKey, strings.Join(messages, ". "), outbubble.WithDisplayBubble())
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}

// errorKeys maps expected domain failures to localized messages.
func errorKeys(cause error) []i18n.Key {
	if errors.Is(cause, domain.ErrAccessDenied) {
		return []i18n.Key{"room.floorplan.error.access_denied"}
	}
	var validation domain.ValidationErrors
	if !errors.As(cause, &validation) {
		return nil
	}
	keys := make([]i18n.Key, len(validation.Codes))
	for index, code := range validation.Codes {
		keys[index] = i18n.Key("room.floorplan.error." + string(code))
	}

	return keys
}
