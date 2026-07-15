// Package interact handles generic furniture use commands.
package interact

import (
	"context"
	"errors"
	"strconv"

	"github.com/niflaot/pixels/internal/command"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	furnitureaccess "github.com/niflaot/pixels/internal/realm/furniture/access"
	decorcmd "github.com/niflaot/pixels/internal/realm/furniture/commands/decor"
	furnituresession "github.com/niflaot/pixels/internal/realm/furniture/commands/session"
	furnitureused "github.com/niflaot/pixels/internal/realm/furniture/events/used"
	"github.com/niflaot/pixels/internal/realm/furniture/interactions"
	essential "github.com/niflaot/pixels/internal/realm/furniture/interactions/essential"
	teleport "github.com/niflaot/pixels/internal/realm/furniture/interactions/teleport"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outstate "github.com/niflaot/pixels/networking/outbound/room/furniture/state"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/zap"
)

const (
	// Name identifies the generic furniture interaction command.
	Name command.Name = "furniture.interact"
)

// Action identifies the protocol interaction request variant.
type Action uint8

const (
	// ActionUse requests the furniture's normal interaction.
	ActionUse Action = iota
	// ActionDice activates a dice roll through Nitro's dedicated packet.
	ActionDice
	// ActionDiceClose closes a settled dice through Nitro's dedicated packet.
	ActionDiceClose
	// ActionColorWheel activates a wall color wheel through Nitro's dedicated packet.
	ActionColorWheel
)

// Command requests interaction with one furniture item.
type Command struct {
	// Handler stores the source connection handler.
	Handler netconn.Context
	// ItemID identifies the used furniture item.
	ItemID int64
	// State stores the client-provided interaction parameter.
	State int32
	// Action identifies the protocol-specific interaction request.
	Action Action
}

// Teleporter starts one paired teleport transition.
type Teleporter interface {
	// Start starts travel from one active source item.
	Start(context.Context, teleport.StartRequest) error
}

// Handler handles generic furniture interaction commands.
type Handler struct {
	// Players stores live player state.
	Players *playerlive.Registry
	// Bindings stores player connection bindings.
	Bindings *binding.Registry
	// Furniture manages durable furniture state.
	Furniture furnitureservice.Manager
	// States changes durable furniture interaction state.
	States furnitureservice.StateUpdater
	// Runtime stores active rooms.
	Runtime *roomlive.Registry
	// Permissions resolves global furniture management authority.
	Permissions permissionservice.Checker
	// Connections stores active network connections.
	Connections *netconn.Registry
	// Events publishes furniture interaction events.
	Events bus.Publisher
	// Translations resolves end-user messages.
	Translations i18n.Translator
	// Behaviors resolves generic state transitions.
	Behaviors *interactions.Registry
	// Teleports coordinates paired furniture travel.
	Teleports Teleporter
	// Essentials coordinates specialized furniture interactions.
	Essentials *essential.Service
	// Decorator handles mannequin wear and background-toner toggle behavior.
	Decorator *decorcmd.Handler
	// Log records malformed durable state and interaction diagnostics.
	Log *zap.Logger
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name {
	return Name
}

// Handle routes one furniture use through its definition behavior.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, err := furnituresession.Player(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	roomID, found := player.CurrentRoom()
	if !found {
		return nil
	}
	active, found := handler.Runtime.Find(roomID)
	if !found {
		return nil
	}
	item, found := active.FurnitureItem(envelope.Command.ItemID)
	if !found {
		return nil
	}
	if !actionMatches(envelope.Command.Action, item.Definition.InteractionType) {
		return nil
	}
	if envelope.Command.Action == ActionDiceClose && handler.Essentials != nil {
		return handler.Essentials.CloseDice(ctx, essential.Request{PlayerID: player.ID(), Room: active, Item: item})
	}
	if teleportType(item.Definition.InteractionType) {
		return handler.startTeleport(ctx, player.ID(), active, item.ID)
	}
	if handler.Decorator != nil {
		handled, decorErr := handler.Decorator.Use(ctx, player, active, item)
		if handled || decorErr != nil {
			return decorErr
		}
	}
	if handler.Essentials != nil {
		handled, essentialErr := handler.Essentials.Use(ctx, essential.Request{PlayerID: player.ID(), Room: active, Item: item})
		if errors.Is(essentialErr, essential.ErrNoRights) {
			return handler.sendNoRights(ctx, envelope.Command.Handler)
		}
		if handled || essentialErr != nil {
			return essentialErr
		}
	}
	behavior, found := handler.Behaviors.Resolve(item.Definition.InteractionType)
	if !found {
		return nil
	}
	allowed, err := furnitureaccess.CanManage(ctx, handler.Permissions, active, player.ID())
	if err != nil {
		return err
	}
	if !allowed {
		return handler.sendNoRights(ctx, envelope.Command.Handler)
	}
	if toggleType(item.Definition.InteractionType) && !validToggleState(item.ExtraData, item.Definition.InteractionModesCount) && handler.Log != nil {
		handler.Log.Warn("furniture interaction state invalid",
			zap.Int64("item_id", item.ID), zap.Int64("room_id", roomID), zap.String("interaction_type", item.Definition.InteractionType),
			zap.String("extra_data", item.ExtraData), zap.Int("modes", item.Definition.InteractionModesCount),
		)
	}
	next, rebuild, commit := behavior.Next(active, item)
	if !commit {
		return nil
	}
	_, err = handler.States.UpdateState(ctx, furnitureservice.StateParams{
		ItemID: item.ID, RoomID: roomID, Expected: item.ExtraData, Next: next,
	})
	if errors.Is(err, furnitureservice.ErrStateConflict) {
		return handler.resync(ctx, active, item.ID, item.Definition.InteractionType)
	}
	if err != nil {
		return err
	}
	updated, err := active.UpdateFurnitureState(item.ID, next, rebuild)
	if err != nil {
		return err
	}
	if err := handler.broadcast(ctx, active, updated.ID, updated.ExtraData); err != nil {
		return err
	}

	return handler.publish(ctx, player.ID(), updated.ID, roomID)
}

// actionMatches validates dedicated protocol packets against their furniture type.
func actionMatches(action Action, interactionType string) bool {
	switch action {
	case ActionDice, ActionDiceClose:
		return interactionType == "dice"
	case ActionColorWheel:
		return interactionType == "colorwheel"
	default:
		return true
	}
}

// startTeleport delegates paired travel without applying furniture management rights.
func (handler Handler) startTeleport(ctx context.Context, playerID int64, active *roomlive.Room, itemID int64) error {
	err := handler.Teleports.Start(ctx, teleport.StartRequest{PlayerID: playerID, Room: active, ItemID: itemID})
	if errors.Is(err, teleport.ErrNotTeleport) || errors.Is(err, teleport.ErrInvalidUse) {
		return nil
	}

	return err
}

// broadcast publishes one compact protocol state update to every occupant.
func (handler Handler) broadcast(ctx context.Context, active *roomlive.Room, itemID int64, value string) error {
	state, err := strconv.Atoi(value)
	if err != nil {
		state = 0
	}
	packet, err := outstate.Encode(itemID, state)
	if err != nil {
		return err
	}

	return broadcast.RoomPacket(ctx, handler.Connections, active, packet, 0)
}

// publish emits one successful generic furniture use event.
func (handler Handler) publish(ctx context.Context, playerID int64, itemID int64, roomID int64) error {
	if handler.Events == nil {
		return nil
	}

	return handler.Events.Publish(ctx, bus.Event{Name: furnitureused.Name, Payload: furnitureused.Payload{
		PlayerID: playerID, ItemID: itemID, RoomID: roomID,
	}})
}

// teleportType reports whether an interaction delegates to paired travel.
func teleportType(value string) bool {
	return value == "teleport" || value == "teleport_tile"
}

// toggleType reports whether an interaction uses generic state cycling.
func toggleType(value string) bool {
	return value == "default" || value == "toggle"
}

// validToggleState reports whether durable state is inside the declared cycle.
func validToggleState(value string, modes int) bool {
	state, err := strconv.Atoi(value)

	return err == nil && state >= 0 && state < modes
}
