package settings

import (
	"context"
	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/control/commands/resolve"
	settingsupdated "github.com/niflaot/pixels/internal/realm/room/control/events/settingsupdated"
	roomsettings "github.com/niflaot/pixels/internal/realm/room/control/settings"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roombroadcast "github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outupdated "github.com/niflaot/pixels/networking/outbound/room/settings/updated"
	"github.com/niflaot/pixels/pkg/bus"
)

// QuickName identifies the focused category and trade-mode command.
const QuickName command.Name = "room.settings.quick"

// QuickCommand contains a focused room settings mutation.
type QuickCommand struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// RoomID identifies the room.
	RoomID int64
	// CategoryID identifies the navigator category.
	CategoryID int64
	// TradeMode stores direct-trade policy.
	TradeMode int32
}

// CommandName returns the stable command name.
func (QuickCommand) CommandName() command.Name { return QuickName }

// QuickHandler persists a focused room settings mutation.
type QuickHandler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores authenticated connection bindings.
	Bindings *binding.Registry
	// Rooms reads and updates room settings.
	Rooms SaveRoomManager
	// Authorize resolves settings management authority.
	Authorize *roomsettings.Authorizer
	// Runtime stores active room projections.
	Runtime *roomlive.Registry
	// Connections projects the update to current occupants.
	Connections *netconn.Registry
	// Events publishes the committed settings update.
	Events bus.Publisher
}

// Handle authorizes and applies category and trade-mode changes.
func (handler QuickHandler) Handle(ctx context.Context, envelope command.Envelope[QuickCommand]) error {
	input := envelope.Command
	player, currentRoomID, err := control.Actor(input.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	if err = control.MatchRoom(currentRoomID, input.RoomID); err != nil {
		return err
	}
	room, found, err := handler.Rooms.FindByID(ctx, input.RoomID)
	if err != nil {
		return err
	}
	if !found {
		return roomservice.ErrRoomNotFound
	}
	if err = handler.Authorize.Authorize(ctx, room, player.ID()); err != nil {
		return err
	}
	allowReserved, err := handler.Authorize.CanManageAny(ctx, player.ID())
	if err != nil {
		return err
	}
	categoryID := &input.CategoryID
	if input.CategoryID <= 0 {
		categoryID = nil
	}
	tradeMode := roommodel.TradeMode(input.TradeMode)
	updated, err := handler.Rooms.Update(ctx, room.ID, room.Version.Version, roomservice.UpdateParams{CategoryID: &categoryID, TradeMode: &tradeMode, AllowReservedTags: allowReserved})
	if err != nil {
		return err
	}
	var active *roomlive.Room
	activeFound := false
	if handler.Runtime != nil {
		active, activeFound = handler.Runtime.Find(room.ID)
	}
	if activeFound {
		active.UpdateCategoryAndTrade(updated.CategoryID, int16(updated.TradeMode))
	}
	response, err := outupdated.Encode(int32(room.ID))
	if err != nil {
		return err
	}
	if activeFound {
		if err = roombroadcast.RoomPacket(ctx, handler.Connections, active, response, 0); err != nil {
			return err
		}
	} else if err = input.Handler.Send(ctx, response); err != nil {
		return err
	}
	if handler.Events == nil {
		return nil
	}
	return handler.Events.Publish(ctx, bus.Event{Name: settingsupdated.Name, Payload: settingsupdated.Payload{RoomID: room.ID, ActorID: player.ID(), Version: updated.Version.Version}})
}
