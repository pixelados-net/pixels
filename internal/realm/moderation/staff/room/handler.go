// Package room adapts moderator-tool room alerts and room overrides.
package room

import (
	"context"
	"errors"
	"strings"

	moderationpolicy "github.com/niflaot/pixels/internal/realm/moderation/policy"
	moderationruntime "github.com/niflaot/pixels/internal/realm/moderation/runtime"
	roomrealm "github.com/niflaot/pixels/internal/realm/room"
	leavecmd "github.com/niflaot/pixels/internal/realm/room/access/commands/leave"
	settingsupdated "github.com/niflaot/pixels/internal/realm/room/control/events/settingsupdated"
	moderationbroadcast "github.com/niflaot/pixels/internal/realm/room/control/moderation/broadcast"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roombroadcast "github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inchange "github.com/niflaot/pixels/networking/inbound/moderation/staff/changeroom"
	inalert "github.com/niflaot/pixels/networking/inbound/moderation/staff/roomalert"
	outupdated "github.com/niflaot/pixels/networking/outbound/room/settings/updated"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	"github.com/niflaot/pixels/pkg/bus"
)

const (
	// alertAction identifies Nitro's room alert action.
	alertAction int32 = 0
	// messageAction identifies Nitro's room caution action.
	messageAction int32 = 3
)

var (
	// errInvalidAction reports an unsupported Nitro room action.
	errInvalidAction = errors.New("invalid moderator room action")
	// errActorOffline reports missing live state for an authenticated moderator.
	errActorOffline = errors.New("moderator live player not found")
	// errActorOutsideRoom reports a room alert without current room presence.
	errActorOutsideRoom = errors.New("moderator is not in a room")
	// errUnauthorized reports a missing room override capability.
	errUnauthorized = errors.New("moderator room override denied")
)

// Handler adapts moderator room packet behavior.
type Handler struct {
	// Context provides shared moderation and room dependencies.
	*moderationruntime.Context
}

// Register installs moderator room packet adapters.
func Register(registry *netconn.HandlerRegistry, runtime *moderationruntime.Context) {
	handler := Handler{Context: runtime}
	_ = registry.Register(inalert.Header, handler.alert)
	_ = registry.Register(inchange.Header, handler.change)
}

// alert broadcasts a moderator message to the actor's current room.
func (handler Handler) alert(connection netconn.Context, packet codec.Packet) error {
	payload, err := inalert.Decode(packet)
	if err != nil {
		return err
	}
	if payload.Action != alertAction && payload.Action != messageAction {
		return errInvalidAction
	}
	actorID, err := handler.authorize(connection)
	if err != nil {
		return err
	}
	actor, found := handler.Players.Find(actorID)
	if !found {
		return errActorOffline
	}
	roomID, found := actor.CurrentRoom()
	if !found {
		return errActorOutsideRoom
	}
	active, found := handler.RoomsLive.Find(roomID)
	if !found {
		return roomlive.ErrRoomNotFound
	}
	message := strings.TrimSpace(handler.Moderation.Sanitize(payload.Message))
	response, err := outalert.Encode(message)
	if err != nil {
		return err
	}
	var deliveryErr error
	for _, occupant := range active.Occupants() {
		deliveryErr = errors.Join(deliveryErr, handler.SendTo(context.Background(), occupant.PlayerID, response))
	}
	return deliveryErr
}

// change applies moderator room settings and optional occupant removal.
func (handler Handler) change(connection netconn.Context, packet codec.Packet) error {
	payload, err := inchange.Decode(packet)
	if err != nil {
		return err
	}
	actorID, err := handler.authorize(connection)
	if err != nil {
		return err
	}
	roomRecord, found, err := handler.Rooms.FindByID(context.Background(), int64(payload.RoomID))
	if err != nil {
		return err
	}
	if !found {
		return roomservice.ErrRoomNotFound
	}
	updated, changed, err := handler.update(roomRecord, payload)
	if err != nil {
		return err
	}
	active, activeFound := handler.RoomsLive.Find(roomRecord.ID)
	if changed && activeFound {
		if err = handler.broadcastUpdate(active, updated.ID); err != nil {
			return err
		}
	}
	if changed && handler.Events != nil {
		err = handler.Events.Publish(context.Background(), bus.Event{Name: settingsupdated.Name, Payload: settingsupdated.Payload{RoomID: updated.ID, ActorID: actorID, Version: updated.Version.Version}})
	}
	if err == nil && payload.KickUsers == 1 && activeFound {
		err = handler.kickOccupants(active, actorID)
	}
	return err
}

// authorize resolves the actor and global room override permission.
func (handler Handler) authorize(connection netconn.Context) (int64, error) {
	actorID, err := handler.Actor(connection)
	if err != nil {
		return 0, err
	}
	allowed, err := handler.Permissions.HasPermission(context.Background(), actorID, moderationpolicy.RoomOverride)
	if err != nil {
		return 0, err
	}
	if !allowed {
		return 0, errUnauthorized
	}
	return actorID, nil
}

// update applies only the selected persistent moderator overrides.
func (handler Handler) update(roomRecord roommodel.Room, payload inchange.Payload) (roommodel.Room, bool, error) {
	params := roomservice.UpdateParams{AllowReservedTags: true}
	changed := false
	if payload.LockDoor == 1 {
		mode := roommodel.DoorModeDoorbell
		params.DoorMode = &mode
		changed = true
	}
	if payload.ChangeTitle == 1 {
		name := handler.Text("moderation.room.inappropriate_name")
		params.Name = &name
		changed = true
	}
	if !changed {
		return roomRecord, false, nil
	}
	updated, err := handler.Rooms.Update(context.Background(), roomRecord.ID, roomRecord.Version.Version, params)
	return updated, true, err
}

// broadcastUpdate refreshes room information for current occupants.
func (handler Handler) broadcastUpdate(active *roomlive.Room, roomID int64) error {
	response, err := outupdated.Encode(int32(roomID))
	if err != nil {
		return err
	}
	return roombroadcast.RoomPacket(context.Background(), handler.Connections, active, response, 0)
}

// kickOccupants removes ordinary occupants while preserving protected users.
func (handler Handler) kickOccupants(active *roomlive.Room, actorID int64) error {
	leaver := leavecmd.Handler{Players: handler.Players, Bindings: handler.Bindings, Runtime: handler.RoomsLive, Connections: handler.Connections, Events: handler.Events}
	notice, err := moderationbroadcast.KickedNotice(handler.Translations)
	if err != nil {
		return err
	}
	var kickErr error
	for _, occupant := range active.Occupants() {
		protected, protectionErr := handler.protected(occupant.PlayerID, actorID)
		kickErr = errors.Join(kickErr, protectionErr)
		if protected || protectionErr != nil {
			continue
		}
		kickErr = errors.Join(kickErr, leaver.ToDesktopThen(context.Background(), occupant.PlayerID, notice))
	}
	return kickErr
}

// protected reports whether a room occupant must survive a mass kick.
func (handler Handler) protected(playerID int64, actorID int64) (bool, error) {
	if playerID == actorID {
		return true, nil
	}
	staff, err := handler.Permissions.HasPermission(context.Background(), playerID, moderationpolicy.ToolAccess)
	if err != nil || staff {
		return staff, err
	}
	return handler.Permissions.HasPermission(context.Background(), playerID, roomrealm.Unkickable)
}
