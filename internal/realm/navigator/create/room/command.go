// Package create creates rooms through the navigator realm.
package create

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	navsession "github.com/niflaot/pixels/internal/realm/navigator/session"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/record/events/created"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outcreated "github.com/niflaot/pixels/networking/outbound/navigator/create/roomcreated"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// Name identifies the navigator room create command.
	Name command.Name = "navigator.create_room"
)

// Command creates a room.
type Command struct {
	// Handler stores the source connection handler.
	Handler netconn.Context
	// RoomName stores the requested room name.
	RoomName string
	// RoomDescription stores the requested room description.
	RoomDescription string
	// ModelName stores the requested room model.
	ModelName string
	// CategoryID stores the requested category id.
	CategoryID int32
	// MaxVisitors stores the requested maximum users.
	MaxVisitors int32
	// TradeType stores the requested trade mode.
	TradeType int32
}

// Handler handles room creation commands.
type Handler struct {
	// Players stores live player state.
	Players *playerlive.Registry
	// Bindings stores player connection bindings.
	Bindings *binding.Registry
	// Rooms creates persistent rooms.
	Rooms RoomManager
	// Events publishes room lifecycle events.
	Events bus.Publisher
	// Translations resolves localized creation feedback.
	Translations i18n.Translator
	// Log records rejected creation attempts.
	Log *zap.Logger
}

// RoomManager combines the focused room creation dependencies.
type RoomManager interface {
	roomservice.Creator
	roomservice.OwnerLister
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name {
	return Name
}

// MarshalLogObject writes safe debug command fields.
func (input Command) MarshalLogObject(encoder zapcore.ObjectEncoder) error {
	encoder.AddString("connection_id", string(input.Handler.ConnectionID))
	encoder.AddString("room_name", input.RoomName)
	encoder.AddString("model_name", input.ModelName)
	encoder.AddInt32("category_id", input.CategoryID)
	encoder.AddInt32("max_visitors", input.MaxVisitors)
	encoder.AddInt32("trade_type", input.TradeType)

	return nil
}

// Handle handles a room creation command.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, _, err := navsession.Player(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	rooms, err := handler.Rooms.ListByOwner(ctx, player.ID())
	if err != nil {
		return err
	}
	if len(rooms) >= roomservice.MaxRoomsPerPlayer {
		return handler.sendLimit(ctx, envelope.Command.Handler)
	}

	room, err := handler.Rooms.Create(ctx, createParams(envelope.Command, player.ID(), player.Username()))
	if err != nil {
		return handler.handleCreateError(ctx, envelope.Command, player.ID(), err)
	}
	if err := handler.publish(ctx, room.ID, player.ID()); err != nil {
		return err
	}

	packet, err := outcreated.Encode(int32(room.ID), room.Name)
	if err != nil {
		return err
	}

	return envelope.Command.Handler.Send(ctx, packet)
}

// createParams maps command input to room service input.
func createParams(input Command, playerID int64, username string) roomservice.CreateParams {
	return roomservice.CreateParams{
		OwnerPlayerID: playerID,
		OwnerName:     username,
		Name:          input.RoomName,
		Description:   input.RoomDescription,
		ModelName:     input.ModelName,
		MaxUsers:      int(input.MaxVisitors),
		CategoryID:    categoryID(input.CategoryID),
		TradeMode:     roommodel.TradeMode(input.TradeType),
	}
}

// categoryID maps protocol category ids to optional records.
func categoryID(id int32) *int64 {
	if id <= 0 {
		return nil
	}

	value := int64(id)

	return &value
}

// publish emits room creation.
func (handler Handler) publish(ctx context.Context, roomID int64, playerID int64) error {
	if handler.Events == nil {
		return nil
	}

	return handler.Events.Publish(ctx, bus.Event{Name: created.Name, Payload: created.Payload{RoomID: roomID, OwnerPlayerID: playerID}})
}
