package decor

import (
	"context"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomdecor "github.com/niflaot/pixels/internal/realm/room/decoration"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	outrefresh "github.com/niflaot/pixels/networking/outbound/inventory/furniture/refresh"
	outremove "github.com/niflaot/pixels/networking/outbound/inventory/furniture/remove"
	outpaint "github.com/niflaot/pixels/networking/outbound/room/paint"
)

// handleSurface consumes one floor, wallpaper, or landscape item.
func (handler Handler) handleSurface(ctx context.Context, player *playerlive.Player, active *roomlive.Room, roomID int64, command Command) error {
	allowed, err := handler.canManage(ctx, active, player.ID())
	if err != nil || !allowed {
		return err
	}
	item, found, err := handler.Furniture.FindItemByID(ctx, command.ItemID)
	if err != nil || !found || !item.InInventory() || item.OwnerPlayerID != player.ID() {
		return err
	}
	definition, found, err := handler.Furniture.FindDefinitionByID(ctx, item.DefinitionID)
	if err != nil || !found {
		return err
	}
	surface := roomdecor.Surface(definition.Name)
	if err = handler.Decoration.ApplySurface(ctx, item.ID, player.ID(), roomID, surface, item.ExtraData); err != nil {
		if decorationSoftError(err) {
			return nil
		}
		return err
	}
	packet, err := outpaint.Encode(string(surface), item.ExtraData)
	if err != nil {
		return err
	}
	if err = handler.broadcast(ctx, active, packet); err != nil {
		return err
	}
	packet, err = outremove.Encode(item.ID)
	if err != nil {
		return err
	}
	if err = command.Handler.Send(ctx, packet); err != nil {
		return err
	}
	packet, err = outrefresh.Encode()
	if err != nil {
		return err
	}
	return command.Handler.Send(ctx, packet)
}
