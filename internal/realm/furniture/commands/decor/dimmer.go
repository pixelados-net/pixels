package decor

import (
	"context"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomdecor "github.com/niflaot/pixels/internal/realm/room/decoration"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	outsettings "github.com/niflaot/pixels/networking/outbound/room/dimmer/settings"
	outwallupdate "github.com/niflaot/pixels/networking/outbound/room/furniture/wallupdate"
)

// handleDimmer handles preset requests, saves, and toggles.
func (handler Handler) handleDimmer(ctx context.Context, player *playerlive.Player, active *roomlive.Room, roomID int64, command Command) error {
	state, found, err := handler.Decoration.LoadDimmer(ctx, roomID)
	if err != nil || !found {
		return err
	}
	if command.Kind == KindDimmerSettings {
		return sendDimmerSettings(ctx, command, state)
	}
	allowed, err := handler.canManage(ctx, active, player.ID())
	if err != nil || !allowed {
		return err
	}
	ownerID := active.Snapshot().OwnerPlayerID
	if command.Kind == KindDimmerSave {
		state, err = handler.Decoration.SaveDimmer(ctx, roomID, ownerID, roomdecor.Preset{ID: command.PresetID, BackgroundOnly: command.Type == 2, Color: command.Color, Brightness: command.First}, command.Apply)
	} else {
		state, err = handler.Decoration.ToggleDimmer(ctx, roomID, ownerID)
	}
	if err != nil {
		if decorationSoftError(err) {
			return nil
		}
		return err
	}
	item, found, err := handler.Furniture.FindItemByID(ctx, state.ItemID)
	if err != nil || !found || item.WallPosition == nil {
		return err
	}
	definition, found, err := handler.Furniture.FindDefinitionByID(ctx, item.DefinitionID)
	if err != nil || !found {
		return err
	}
	packet, err := outwallupdate.Encode(item.ID, definition.SpriteID, *item.WallPosition, state.ExtraData, 0, item.OwnerPlayerID)
	if err != nil {
		return err
	}
	if err = handler.broadcast(ctx, active, packet); err != nil {
		return err
	}
	return sendDimmerSettings(ctx, command, state)
}

// sendDimmerSettings sends the three preset slots to the actor.
func sendDimmerSettings(ctx context.Context, command Command, state roomdecor.DimmerState) error {
	packet, err := outsettings.Encode(state.Presets)
	if err != nil {
		return err
	}
	return command.Handler.Send(ctx, packet)
}
