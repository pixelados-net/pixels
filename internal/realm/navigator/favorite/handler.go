// Package favorite owns the persistent room favorite lifecycle.
package favorite

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/permission"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	navservice "github.com/niflaot/pixels/internal/realm/navigator/core"
	favoritechanged "github.com/niflaot/pixels/internal/realm/navigator/favorite/events/changed"
	navmodel "github.com/niflaot/pixels/internal/realm/navigator/record"
	navsession "github.com/niflaot/pixels/internal/realm/navigator/session"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomrights "github.com/niflaot/pixels/internal/realm/room/control/rights"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inadd "github.com/niflaot/pixels/networking/inbound/navigator/favorite/add"
	inremove "github.com/niflaot/pixels/networking/inbound/navigator/favorite/remove"
	outchanged "github.com/niflaot/pixels/networking/outbound/navigator/favorite/changed"
	outfavorites "github.com/niflaot/pixels/networking/outbound/navigator/favorite/list"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
)

var (
	// Unlimited bypasses the ordinary per-player favorite limit.
	Unlimited = permission.RegisterNode("navigator.favorite.unlimited", "")
)

// Handler adapts Nitro favorite packets to Navigator persistence.
type Handler struct {
	// Navigator coordinates favorite persistence.
	Navigator navservice.Manager
	// Players stores live navigator viewers.
	Players *playerlive.Registry
	// Bindings resolves authenticated players.
	Bindings *binding.Registry
	// Events publishes committed favorite changes.
	Events bus.Publisher
	// Rooms validates favorite targets.
	Rooms roomservice.Manager
	// Rights resolves invisible-room visibility.
	Rights roomrights.Manager
	// Permissions resolves the configured favorite quota bypass.
	Permissions permissionservice.Checker
	// Limit bounds ordinary player favorites.
	Limit int32
	// Translations resolves expected favorite feedback.
	Translations i18n.Translator
}

// RegisterHandlers installs favorite add and remove adapters.
func RegisterHandlers(registry *netconn.HandlerRegistry, handler Handler) {
	if registry == nil {
		return
	}
	_ = registry.Register(inadd.Header, handler.add)
	_ = registry.Register(inremove.Header, handler.remove)
}

// add validates, persists, and refreshes one favorite.
func (handler Handler) add(connection netconn.Context, packet codec.Packet) error {
	payload, err := inadd.Decode(packet)
	if err != nil {
		return err
	}
	return handler.change(connection, int64(payload.RoomID), true)
}

// remove persists and refreshes one favorite removal.
func (handler Handler) remove(connection netconn.Context, packet codec.Packet) error {
	payload, err := inremove.Decode(packet)
	if err != nil {
		return err
	}
	return handler.change(connection, int64(payload.RoomID), false)
}

// change executes one favorite mutation and sends both incremental and complete state.
func (handler Handler) change(connection netconn.Context, roomID int64, added bool) error {
	player, _, err := navsession.Player(connection, handler.Bindings, handler.Players)
	if err != nil || handler.Navigator == nil {
		return err
	}
	if added {
		room, found, findErr := handler.Rooms.FindByID(context.Background(), roomID)
		if findErr != nil {
			return findErr
		}
		if !found || room.DoorMode == roommodel.DoorModeInvisible && room.OwnerPlayerID != player.ID() && !handler.hasRights(roomID, player.ID()) {
			return handler.alert(connection, "navigator.favorite.room_unavailable")
		}
		current, listErr := handler.Navigator.ListFavoriteRoomIDs(context.Background(), player.ID())
		if listErr != nil {
			return listErr
		}
		for _, existing := range current {
			if existing == roomID {
				return handler.refresh(connection, player.ID(), roomID, true, current)
			}
		}
		unlimited := false
		if handler.Permissions != nil {
			unlimited, err = handler.Permissions.HasPermission(context.Background(), player.ID(), Unlimited)
			if err != nil {
				return err
			}
		}
		if !unlimited && len(current) >= int(handler.limit()) {
			return handler.alert(connection, "navigator.favorite.limit")
		}
		err = handler.Navigator.AddFavorite(context.Background(), player.ID(), roomID, handler.limit(), unlimited)
	} else {
		err = handler.Navigator.RemoveFavorite(context.Background(), player.ID(), roomID)
	}
	if err != nil {
		if errors.Is(err, navmodel.ErrFavoriteUnavailable) {
			return handler.alert(connection, "navigator.favorite.room_unavailable")
		}
		return err
	}
	roomIDs, err := handler.Navigator.ListFavoriteRoomIDs(context.Background(), player.ID())
	if err != nil {
		return err
	}
	return handler.refresh(connection, player.ID(), roomID, added, roomIDs)
}

// alert sends localized expected favorite feedback without disconnecting the session.
func (handler Handler) alert(connection netconn.Context, key i18n.Key) error {
	message := string(key)
	if handler.Translations != nil {
		message = handler.Translations.Default(key)
	}
	packet, err := outalert.Encode(message)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), packet)
}

// hasRights reports privacy-safe invisible room visibility.
func (handler Handler) hasRights(roomID int64, playerID int64) bool {
	if handler.Rights == nil {
		return false
	}
	allowed, err := handler.Rights.HasRights(context.Background(), roomID, playerID)
	return err == nil && allowed
}

// refresh sends incremental and complete favorite state after persistence.
func (handler Handler) refresh(connection netconn.Context, playerID int64, roomID int64, added bool, roomIDs []int64) error {
	changed, err := outchanged.Encode(int32(roomID), added)
	if err != nil {
		return err
	}
	if err = connection.Send(context.Background(), changed); err != nil {
		return err
	}
	values := make([]int32, len(roomIDs))
	for index, value := range roomIDs {
		values[index] = int32(value)
	}
	list, err := outfavorites.Encode(handler.limit(), values)
	if err != nil {
		return err
	}
	if err = connection.Send(context.Background(), list); err != nil {
		return err
	}
	if handler.Events == nil {
		return nil
	}
	return handler.Events.Publish(context.Background(), bus.Event{Name: favoritechanged.Name, Payload: favoritechanged.Payload{PlayerID: playerID, RoomID: roomID, Added: added}})
}

// limit returns the configured ordinary favorite capacity.
func (handler Handler) limit() int32 {
	if handler.Limit <= 0 {
		return 30
	}
	return handler.Limit
}
