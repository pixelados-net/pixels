// Package present opens wrapped gift furniture.
package present

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/command"
	furnituresession "github.com/niflaot/pixels/internal/realm/furniture/commands/session"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roomfurniture "github.com/niflaot/pixels/internal/realm/room/world/items"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outopened "github.com/niflaot/pixels/networking/outbound/inventory/furniture/present/opened"
	outremove "github.com/niflaot/pixels/networking/outbound/room/furniture/remove"
	"go.uber.org/zap"
)

const (
	// Name identifies the furniture present open command.
	Name command.Name = "furniture.present.open"
)

// Command opens a placed gift.
type Command struct {
	// Handler stores the source connection handler.
	Handler netconn.Context
	// ItemID identifies the placed gift furniture item.
	ItemID int64
}

// Handler handles present open commands.
type Handler struct {
	// Players stores live player state.
	Players *playerlive.Registry
	// Bindings stores player connection bindings.
	Bindings *binding.Registry
	// Furniture manages placed and inventory furniture records.
	Furniture furnitureservice.Manager
	// Runtime stores active rooms.
	Runtime *roomlive.Registry
	// Connections stores active network connections.
	Connections *netconn.Registry
	// Log records rejected open attempts.
	Log *zap.Logger
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name {
	return Name
}

// Handle handles a present open command.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, err := furnituresession.Player(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	roomID, found := player.CurrentRoom()
	if !found || handler.Furniture == nil || handler.Runtime == nil {
		return nil
	}
	active, found := handler.Runtime.Find(roomID)
	if !found {
		return nil
	}

	opener, ok := handler.Furniture.(furnitureservice.GiftOpener)
	if !ok {
		return furnitureservice.ErrGiftOpenerUnavailable
	}

	opened, err := opener.OpenGift(ctx, furnitureservice.OpenGiftParams{
		ItemID: envelope.Command.ItemID, ActorPlayerID: player.ID(), RoomID: roomID,
	})
	if err != nil {
		return handler.handleOpenError(envelope.Command, err)
	}
	definition, found, err := handler.Furniture.FindDefinitionByID(ctx, opened.DefinitionID)
	if err != nil || !found {
		return err
	}
	if err := handler.reloadRuntime(active, opened, definition); err != nil {
		return err
	}
	if err := handler.broadcastRefresh(ctx, active, opened, definition, player.Username()); err != nil {
		return err
	}

	return handler.sendOpened(ctx, envelope.Command.Handler, opened, definition)
}

// handleOpenError handles expected present-open rejections without disconnecting the session.
func (handler Handler) handleOpenError(command Command, err error) error {
	if expectedOpenError(err) {
		if handler.Log != nil {
			handler.Log.Debug("present open rejected", zap.Int64("item_id", command.ItemID), zap.Error(err))
		}
		return nil
	}

	return err
}

// expectedOpenError reports whether an open failure is a user action rejection.
func expectedOpenError(err error) bool {
	return errors.Is(err, furnitureservice.ErrItemNotFound) ||
		errors.Is(err, furnitureservice.ErrNotItemOwner) ||
		errors.Is(err, furnitureservice.ErrItemNotInRoom) ||
		errors.Is(err, furnitureservice.ErrItemNotGift) ||
		errors.Is(err, furnitureservice.ErrInvalidItemID)
}

// reloadRuntime replaces the room world's furniture snapshot.
func (handler Handler) reloadRuntime(active *roomlive.Room, item furnituremodel.Item, definition furnituremodel.Definition) error {
	worldItem, ok, err := roomfurniture.ToWorldItem(item, map[int64]furnituremodel.Definition{definition.ID: definition})
	if err != nil || !ok {
		return err
	}
	_, err = active.ReloadFurniture(item.ID, &worldItem)

	return err
}

// broadcastRefresh removes the wrapper object and adds the opened furniture.
func (handler Handler) broadcastRefresh(ctx context.Context, active *roomlive.Room, item furnituremodel.Item, definition furnituremodel.Definition, ownerName string) error {
	if handler.Connections == nil {
		return nil
	}
	remove, err := outremove.Encode(item.ID, item.OwnerPlayerID)
	if err != nil {
		return err
	}
	if err := broadcast.RoomPacket(ctx, handler.Connections, active, remove, 0); err != nil {
		return err
	}
	add, err := addPacket(item, definition, ownerName)
	if err != nil {
		return err
	}

	return broadcast.RoomPacket(ctx, handler.Connections, active, add, 0)
}

// sendOpened sends Nitro's present-opened product response.
func (handler Handler) sendOpened(ctx context.Context, connection netconn.Context, item furnituremodel.Item, definition furnituremodel.Definition) error {
	packet, err := outopened.Encode(int32(definition.SpriteID), definition.Name, item.ID, true)
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}
