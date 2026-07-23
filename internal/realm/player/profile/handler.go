package profile

import (
	"context"
	"errors"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	respectgranted "github.com/niflaot/pixels/internal/realm/player/profile/events/respectgranted"
	profileupdated "github.com/niflaot/pixels/internal/realm/player/profile/events/updated"
	roombroadcast "github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	ininfo "github.com/niflaot/pixels/networking/inbound/user/info/request"
	infigure "github.com/niflaot/pixels/networking/inbound/user/profile/figure"
	inmotto "github.com/niflaot/pixels/networking/inbound/user/profile/motto"
	inrespect "github.com/niflaot/pixels/networking/inbound/user/profile/respect"
	intags "github.com/niflaot/pixels/networking/inbound/user/profile/tags"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	outfigure "github.com/niflaot/pixels/networking/outbound/user/profile/figure"
	outrespect "github.com/niflaot/pixels/networking/outbound/user/profile/respect"
	outtags "github.com/niflaot/pixels/networking/outbound/user/profile/tags"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/zap"
)

// Handler adapts Nitro public-profile packets to authoritative domain behavior.
type Handler struct {
	// Service coordinates durable profile mutations.
	Service *Service
	// Bindings resolves authenticated connections.
	Bindings *binding.Registry
	// Players stores live player projections.
	Players *playerlive.Registry
	// Rooms resolves active room generations.
	Rooms *roomlive.Registry
	// Connections sends room projections.
	Connections *netconn.Registry
	// Events publishes committed public profile changes.
	Events bus.Publisher
	// Translations resolves expected validation feedback.
	Translations i18n.Translator
	// Log records recoverable profile mutation failures.
	Log *zap.Logger
}

// RegisterHandlers installs public-profile packet adapters.
func RegisterHandlers(registry *netconn.HandlerRegistry, handler Handler) {
	if registry == nil {
		return
	}
	_ = registry.Register(ininfo.Header, handler.info)
	_ = registry.Register(infigure.Header, handler.figure)
	_ = registry.Register(inmotto.Header, handler.motto)
	_ = registry.Register(intags.Header, handler.tags)
	_ = registry.Register(inrespect.Header, handler.respect)
}

// info sends one real live USER_INFO snapshot.
func (handler Handler) info(connection netconn.Context, packet codec.Packet) error {
	if _, err := ininfo.Decode(packet); err != nil {
		return err
	}
	player, playerID, found := handler.player(connection)
	if !found {
		return nil
	}
	if handler.Service != nil {
		state, err := handler.Service.RespectState(context.Background(), playerID)
		if err != nil {
			return err
		}
		player.SetRespect(state.Received, state.UserRemaining, state.PetRemaining)
	}
	return sendInfo(connection, player.Snapshot())
}

// figure commits and projects one avatar replacement.
func (handler Handler) figure(connection netconn.Context, packet codec.Packet) error {
	payload, err := infigure.Decode(packet)
	if err != nil {
		return err
	}
	player, playerID, found := handler.player(connection)
	if !found || handler.Service == nil {
		return nil
	}
	record, err := handler.Service.UpdateFigure(context.Background(), playerID, payload.Gender, payload.Figure)
	if err != nil {
		if errors.Is(err, ErrInvalidFigure) {
			return handler.alert(connection, "user.profile.figure_invalid")
		}
		return err
	}
	player.SetProfile(record.Profile.Look, record.Profile.Gender, record.Profile.Motto)
	response, err := outfigure.Encode(record.Profile.Look, string(record.Profile.Gender))
	if err != nil {
		return err
	}
	if err = connection.Send(context.Background(), response); err != nil {
		return err
	}
	if err = handler.project(playerID, record.Profile.Look, string(record.Profile.Gender), record.Profile.Motto); err != nil {
		return err
	}
	return handler.publishUpdated(playerID, true, false)
}

// motto commits and projects one public motto replacement.
func (handler Handler) motto(connection netconn.Context, packet codec.Packet) error {
	payload, err := inmotto.Decode(packet)
	if err != nil {
		return err
	}
	player, playerID, found := handler.player(connection)
	if !found || handler.Service == nil {
		return nil
	}
	record, err := handler.Service.UpdateMotto(context.Background(), playerID, payload.Motto)
	if err != nil {
		if errors.Is(err, ErrInvalidMotto) {
			return handler.alert(connection, "user.profile.motto_invalid")
		}
		return handler.mutationFailed(connection, playerID, "motto", err)
	}
	player.SetProfile(record.Profile.Look, record.Profile.Gender, record.Profile.Motto)
	if err = handler.project(playerID, record.Profile.Look, string(record.Profile.Gender), record.Profile.Motto); err != nil {
		return handler.mutationFailed(connection, playerID, "motto_projection", err)
	}
	if err = handler.publishUpdated(playerID, false, true); err != nil {
		return handler.mutationFailed(connection, playerID, "motto_event", err)
	}
	return nil
}

// tags resolves a room-local unit and returns its ordered public tags.
func (handler Handler) tags(connection netconn.Context, packet codec.Packet) error {
	payload, err := intags.Decode(packet)
	if err != nil {
		return err
	}
	_, actorID, found := handler.player(connection)
	if !found || handler.Rooms == nil || handler.Service == nil {
		return nil
	}
	active, found := handler.Rooms.FindByPlayer(actorID)
	if !found {
		return nil
	}
	unit, found := active.UnitByID(int64(payload.RoomUnitID))
	if !found || unit.PlayerID <= 0 {
		return nil
	}
	values, err := handler.Service.Tags(context.Background(), unit.PlayerID)
	if err != nil {
		return err
	}
	response, err := outtags.Encode(payload.RoomUnitID, values)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// respect validates shared-room presence before committing one grant.
func (handler Handler) respect(connection netconn.Context, packet codec.Packet) error {
	payload, err := inrespect.Decode(packet)
	if err != nil {
		return err
	}
	actor, actorID, found := handler.player(connection)
	if !found || handler.Rooms == nil || handler.Service == nil {
		return nil
	}
	active, found := handler.Rooms.FindByPlayer(actorID)
	if !found {
		return handler.alert(connection, "user.profile.respect_same_user")
	}
	if _, found = active.Occupant(int64(payload.TargetPlayerID)); !found {
		return handler.alert(connection, "user.profile.respect_same_user")
	}
	result, err := handler.Service.GrantRespect(context.Background(), actorID, int64(payload.TargetPlayerID))
	if err != nil {
		switch {
		case errors.Is(err, ErrRespectExhausted):
			return handler.alert(connection, "user.profile.respect_exhausted")
		case errors.Is(err, ErrRespectAlreadyGranted):
			return handler.alert(connection, "user.profile.respect_duplicate")
		case errors.Is(err, ErrRespectThrottled):
			return handler.alert(connection, "user.profile.respect_throttled")
		case errors.Is(err, ErrRespectNotAllowed):
			return handler.alert(connection, "user.profile.respect_same_user")
		}
		return err
	}
	actor.SetRespect(actor.Snapshot().RespectsReceived, result.Remaining, actor.Snapshot().RespectsPetRemaining)
	if handler.Events != nil {
		_ = handler.Events.Publish(context.Background(), bus.Event{Name: respectgranted.Name, Payload: respectgranted.Payload{ActorPlayerID: actorID, TargetPlayerID: int64(payload.TargetPlayerID)}})
	}
	response, err := outrespect.Encode(payload.TargetPlayerID, result.TotalReceived)
	if err != nil {
		return err
	}
	return roombroadcast.RoomPacket(context.Background(), handler.Connections, active, response, 0)
}

// alert sends localized expected profile feedback without disconnecting the session.
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

// player resolves one authenticated live player without allocating.
func (handler Handler) player(connection netconn.Context) (*playerlive.Player, int64, bool) {
	if handler.Bindings == nil || handler.Players == nil {
		return nil, 0, false
	}
	current, found := handler.Bindings.FindByConnection(binding.ConnectionKey{Kind: connection.ConnectionKind, ID: connection.ConnectionID})
	if !found {
		return nil, 0, false
	}
	player, found := handler.Players.Find(current.PlayerID)
	return player, current.PlayerID, found
}

// project refreshes the active room unit and observers after a committed mutation.
func (handler Handler) project(playerID int64, figure string, gender string, motto string) error {
	if handler.Rooms == nil {
		return nil
	}
	active, found := handler.Rooms.FindByPlayer(playerID)
	if !found || !active.UpdateOccupantProfile(playerID, figure, gender, motto) {
		return nil
	}
	return roombroadcast.RoomSpawn(context.Background(), handler.Connections, active, playerID, 0)
}

// publishUpdated notifies bounded profile observers after durable commit.
func (handler Handler) publishUpdated(playerID int64, figure bool, motto bool) error {
	if handler.Events == nil {
		return nil
	}
	return handler.Events.Publish(context.Background(), bus.Event{Name: profileupdated.Name, Payload: profileupdated.Payload{PlayerID: playerID, Figure: figure, Motto: motto}})
}
