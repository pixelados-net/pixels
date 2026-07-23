package wired

import (
	"context"
	"time"

	"github.com/niflaot/pixels/internal/command"
	successpacket "github.com/niflaot/pixels/networking/outbound/furniture/wired/save/success"
)

// HandleSnapshot captures authoritative selected furniture state.
func (handler Handler) HandleSnapshot(ctx context.Context, envelope command.Envelope[SnapshotCommand]) error {
	player, active, roomID, err := handler.actor(envelope.Command.Handler)
	if err != nil {
		return err
	}
	if err = handler.authorize(ctx, player.ID(), active); err != nil {
		return handler.sendFailure(ctx, envelope.Command.Handler, "room.wired.save.no_rights")
	}
	if _, err = handler.Store.Capture(ctx, roomID, envelope.Command.ItemID); err != nil {
		return handler.sendFailure(ctx, envelope.Command.Handler, "room.wired.snapshot.blocked")
	}
	if err = handler.Engine.Reload(ctx, roomID, time.Now()); err != nil {
		return handler.sendFailure(ctx, envelope.Command.Handler, "room.wired.save.technical")
	}
	packet, err := successpacket.Encode()
	if err != nil {
		return err
	}
	return envelope.Command.Handler.Send(ctx, packet)
}
