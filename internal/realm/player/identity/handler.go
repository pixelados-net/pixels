package identity

import (
	"context"
	"errors"

	namechanged "github.com/niflaot/pixels/internal/realm/player/identity/events/namechanged"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roombroadcast "github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inchange "github.com/niflaot/pixels/networking/inbound/user/name/change"
	incheck "github.com/niflaot/pixels/networking/inbound/user/name/check"
	outchange "github.com/niflaot/pixels/networking/outbound/user/name/change"
	outcheck "github.com/niflaot/pixels/networking/outbound/user/name/check"
	outroomname "github.com/niflaot/pixels/networking/outbound/user/room/name"
	"github.com/niflaot/pixels/pkg/bus"
)

// Handler adapts Nitro name checks and commits to identity behavior.
type Handler struct {
	// Service coordinates reservations and renames.
	Service *Service
	// Bindings resolves authenticated players.
	Bindings *binding.Registry
	// Players stores live identity projections.
	Players *playerlive.Registry
	// Rooms resolves active room generations.
	Rooms *roomlive.Registry
	// Connections sends room identity projections.
	Connections *netconn.Registry
	// Events publishes committed identity changes.
	Events bus.Publisher
}

// RegisterHandlers installs identity packet adapters.
func RegisterHandlers(registry *netconn.HandlerRegistry, handler Handler) {
	if registry == nil {
		return
	}
	_ = registry.Register(incheck.Header, handler.check)
	_ = registry.Register(inchange.Header, handler.change)
}

// check validates and reserves one candidate before returning Nitro's result.
func (handler Handler) check(connection netconn.Context, packet codec.Packet) error {
	payload, err := incheck.Decode(packet)
	if err != nil {
		return err
	}
	_, playerID, found := handler.player(connection)
	if !found || handler.Service == nil {
		return nil
	}
	result, err := handler.Service.Check(context.Background(), playerID, payload.Username)
	if err != nil {
		return err
	}
	response, err := outcheck.Encode(result.Code, result.Username, result.Suggestions)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// change consumes the reservation and projects one committed rename.
func (handler Handler) change(connection netconn.Context, packet codec.Packet) error {
	payload, err := inchange.Decode(packet)
	if err != nil {
		return err
	}
	player, playerID, found := handler.player(connection)
	if !found || handler.Service == nil {
		return nil
	}
	result, err := handler.Service.Rename(context.Background(), playerID, payload.Username)
	if err != nil {
		code := resultCode(err)
		response, encodeErr := outchange.Encode(code, payload.Username, nil)
		if encodeErr != nil {
			return encodeErr
		}
		return connection.Send(context.Background(), response)
	}
	player.SetUsername(result.NewUsername, false)
	response, err := outchange.Encode(ResultAvailable, result.NewUsername, nil)
	if err != nil {
		return err
	}
	if err = connection.Send(context.Background(), response); err != nil {
		return err
	}
	if err = handler.projectRoom(playerID, result.NewUsername); err != nil {
		return err
	}
	if handler.Events == nil {
		return nil
	}
	return handler.Events.Publish(context.Background(), bus.Event{Name: namechanged.Name, Payload: namechanged.Payload{PlayerID: playerID, OldUsername: result.OldUsername, NewUsername: result.NewUsername}})
}

// projectRoom updates one room-local identity and broadcasts its exact unit id.
func (handler Handler) projectRoom(playerID int64, username string) error {
	if handler.Rooms == nil {
		return nil
	}
	active, found := handler.Rooms.FindByPlayer(playerID)
	if !found || !active.UpdateOccupantName(playerID, username) {
		return nil
	}
	unit, found := active.Unit(playerID)
	if !found {
		return nil
	}
	packet, err := outroomname.Encode(int32(playerID), int32(unit.UnitID), username)
	if err != nil {
		return err
	}
	return roombroadcast.RoomPacket(context.Background(), handler.Connections, active, packet, 0)
}

// player resolves one authenticated live player without allocations.
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

// resultCode maps domain failures to stable Nitro result codes.
func resultCode(err error) int32 {
	if errors.Is(err, ErrRenameDisabled) {
		return ResultDisabled
	}
	if errors.Is(err, ErrUsernameTaken) || errors.Is(err, ErrReservationMissing) {
		return ResultTaken
	}
	return ResultInvalid
}
