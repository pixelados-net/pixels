package decor

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
)

// handleToner validates and saves HSL values from the background-color widget.
func (handler Handler) handleToner(ctx context.Context, player *playerlive.Player, active *roomlive.Room, roomID int64, command Command) error {
	allowed, err := handler.canManage(ctx, active, player.ID())
	if err != nil || !allowed {
		return err
	}
	if !eightBit(command.First) || !eightBit(command.Second) || !eightBit(command.Third) {
		return nil
	}
	item, found, err := handler.Furniture.FindItemByID(ctx, command.ItemID)
	if err != nil || !found || item.RoomID == nil || *item.RoomID != roomID {
		return err
	}
	definition, found, err := handler.Furniture.FindDefinitionByID(ctx, item.DefinitionID)
	if err != nil || !found || definition.InteractionType != "background_toner" {
		return err
	}
	enabled, _, _, _, valid := parseToner(item.ExtraData)
	if !valid {
		enabled = 0
	}
	updated, err := handler.furnitureState(ctx, item, roomID, fmt.Sprintf("%d:%d:%d:%d", enabled, command.First, command.Second, command.Third))
	if err != nil {
		return err
	}
	active.SetFurnitureExtraData(updated.ID, updated.ExtraData)
	return handler.broadcastFloorUpdate(ctx, active, updated, definition)
}

// toggleToner changes only the persisted enabled component.
func (handler Handler) toggleToner(ctx context.Context, active *roomlive.Room, runtimeItem worldfurniture.Item) error {
	item, found, err := handler.Furniture.FindItemByID(ctx, runtimeItem.ID)
	if err != nil || !found {
		return err
	}
	definition, found, err := handler.Furniture.FindDefinitionByID(ctx, item.DefinitionID)
	if err != nil || !found {
		return err
	}
	enabled, hue, saturation, lightness, valid := parseToner(item.ExtraData)
	if !valid {
		enabled, hue, saturation, lightness = 0, 0, 0, 0
	}
	if enabled == 0 {
		enabled = 1
	} else {
		enabled = 0
	}
	updated, err := handler.furnitureState(ctx, item, active.ID(), fmt.Sprintf("%d:%d:%d:%d", enabled, hue, saturation, lightness))
	if err != nil {
		return err
	}
	active.SetFurnitureExtraData(updated.ID, updated.ExtraData)
	return handler.broadcastFloorUpdate(ctx, active, updated, definition)
}

// parseToner parses enabled, hue, saturation, and lightness state.
func parseToner(value string) (int32, int32, int32, int32, bool) {
	parts := strings.Split(value, ":")
	if len(parts) != 4 {
		return 0, 0, 0, 0, false
	}
	values := [4]int32{}
	for index, part := range parts {
		parsed, err := strconv.ParseInt(part, 10, 32)
		if err != nil || !eightBit(int32(parsed)) {
			return 0, 0, 0, 0, false
		}
		values[index] = int32(parsed)
	}
	if values[0] > 1 {
		return 0, 0, 0, 0, false
	}
	return values[0], values[1], values[2], values[3], true
}

// eightBit reports whether one toner component is valid.
func eightBit(value int32) bool { return value >= 0 && value <= 255 }
