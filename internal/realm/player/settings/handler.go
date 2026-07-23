package settings

import (
	"context"
	"errors"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	incamerafollow "github.com/niflaot/pixels/networking/inbound/user/settings/camerafollow"
	inhomeroom "github.com/niflaot/pixels/networking/inbound/user/settings/homeroom"
	inoldchat "github.com/niflaot/pixels/networking/inbound/user/settings/oldchat"
	involume "github.com/niflaot/pixels/networking/inbound/user/settings/volume"
	outalert "github.com/niflaot/pixels/networking/outbound/session/alert"
	outhomeroom "github.com/niflaot/pixels/networking/outbound/user/settings/homeroom"
	"github.com/niflaot/pixels/pkg/i18n"
)

// Handler adapts Nitro settings mutations to durable player settings.
type Handler struct {
	// Service validates and persists settings.
	Service *Service
	// Writer coalesces client-owned settings writes.
	Writer *Writer
	// Bindings resolves authenticated connections.
	Bindings *binding.Registry
	// Players stores live settings projections.
	Players *playerlive.Registry
	// Rooms resolves durable home-room candidates.
	Rooms RoomFinder
	// Rights resolves invisible room eligibility.
	Rights RightsChecker
	// Translations resolves expected validation feedback.
	Translations i18n.Translator
}

// RoomFinder resolves one durable room without exposing persistence details.
type RoomFinder interface {
	// FindByID returns one active room record.
	FindByID(context.Context, int64) (roommodel.Room, bool, error)
}

// RightsChecker resolves explicit room rights.
type RightsChecker interface {
	// HasRights reports whether a player holds explicit rights.
	HasRights(context.Context, int64, int64) (bool, error)
}

// RegisterHandlers installs settings mutation adapters.
func RegisterHandlers(registry *netconn.HandlerRegistry, handler Handler) {
	if registry == nil {
		return
	}
	_ = registry.Register(involume.Header, handler.volume)
	_ = registry.Register(inoldchat.Header, handler.oldChat)
	_ = registry.Register(inhomeroom.Header, handler.homeRoom)
	_ = registry.Register(incamerafollow.Header, handler.cameraFollow)
}

// volume validates, persists, and projects one volume replacement.
func (handler Handler) volume(connection netconn.Context, packet codec.Packet) error {
	payload, err := involume.Decode(packet)
	if err != nil {
		return err
	}
	player, playerID, found := handler.player(connection)
	if !found || handler.Service == nil {
		return nil
	}
	if !validVolume(payload.System) || !validVolume(payload.Furniture) || !validVolume(payload.Trax) {
		return handler.alert(connection, "user.settings.volume_invalid")
	}
	if handler.Writer != nil && handler.Writer.EnqueueVolume(playerID, payload.System, payload.Furniture, payload.Trax) {
		snapshot := player.Snapshot()
		player.SetClientSettings(payload.System, payload.Furniture, payload.Trax, snapshot.OldChat, snapshot.CameraFollowBlocked, snapshot.SafetyLocked)
		return nil
	}
	record, err := handler.Service.SetVolume(context.Background(), playerID, payload.System, payload.Furniture, payload.Trax)
	if err != nil {
		if errors.Is(err, ErrInvalidVolume) {
			return handler.alert(connection, "user.settings.volume_invalid")
		}
		return err
	}
	applyRecord(player, record)
	return nil
}

// oldChat validates, persists, and projects legacy chat selection.
func (handler Handler) oldChat(connection netconn.Context, packet codec.Packet) error {
	payload, err := inoldchat.Decode(packet)
	if err != nil {
		return err
	}
	player, playerID, found := handler.player(connection)
	if !found || handler.Service == nil {
		return nil
	}
	if handler.Writer != nil && handler.Writer.EnqueueOldChat(playerID, payload.OldChat) {
		snapshot := player.Snapshot()
		player.SetClientSettings(snapshot.VolumeSystem, snapshot.VolumeFurniture, snapshot.VolumeTrax, payload.OldChat, snapshot.CameraFollowBlocked, snapshot.SafetyLocked)
		return nil
	}
	record, err := handler.Service.SetOldChat(context.Background(), playerID, payload.OldChat)
	if err != nil {
		return err
	}
	applyRecord(player, record)
	return nil
}

// cameraFollow persists and projects camera-follow privacy.
func (handler Handler) cameraFollow(connection netconn.Context, packet codec.Packet) error {
	payload, err := incamerafollow.Decode(packet)
	if err != nil {
		return err
	}
	player, playerID, found := handler.player(connection)
	if !found || handler.Service == nil {
		return nil
	}
	if handler.Writer != nil && handler.Writer.EnqueueCameraFollowBlocked(playerID, payload.CameraFollowBlocked) {
		snapshot := player.Snapshot()
		player.SetClientSettings(snapshot.VolumeSystem, snapshot.VolumeFurniture, snapshot.VolumeTrax, snapshot.OldChat, payload.CameraFollowBlocked, snapshot.SafetyLocked)
		return nil
	}
	record, err := handler.Service.SetCameraFollowBlocked(context.Background(), playerID, payload.CameraFollowBlocked)
	if err != nil {
		return err
	}
	applyRecord(player, record)
	return nil
}

// homeRoom validates, persists, projects, and confirms home-room selection.
func (handler Handler) homeRoom(connection netconn.Context, packet codec.Packet) error {
	payload, err := inhomeroom.Decode(packet)
	if err != nil {
		return err
	}
	player, playerID, found := handler.player(connection)
	if !found || handler.Service == nil {
		return nil
	}
	var roomID *int64
	if payload.RoomID > 0 {
		value := int64(payload.RoomID)
		if err := handler.validateHomeRoom(playerID, value); err != nil {
			if errors.Is(err, ErrInvalidHomeRoom) {
				return handler.alert(connection, "user.settings.home_room_invalid")
			}
			return err
		}
		roomID = &value
	}
	if err := handler.Service.SetHomeRoom(context.Background(), playerID, roomID); err != nil {
		return err
	}
	player.SetHomeRoom(roomID)
	response, err := outhomeroom.Encode(payload.RoomID, 0)
	if err != nil {
		return err
	}
	return connection.Send(context.Background(), response)
}

// alert sends localized expected settings feedback without disconnecting the session.
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

// validateHomeRoom verifies visibility without bypassing normal room policy.
func (handler Handler) validateHomeRoom(playerID int64, roomID int64) error {
	if handler.Rooms == nil {
		return ErrInvalidHomeRoom
	}
	room, found, err := handler.Rooms.FindByID(context.Background(), roomID)
	if err != nil {
		return err
	}
	if !found {
		return ErrInvalidHomeRoom
	}
	if room.DoorMode != roommodel.DoorModeInvisible || room.OwnerPlayerID == playerID {
		return nil
	}
	if handler.Rights == nil {
		return ErrInvalidHomeRoom
	}
	allowed, err := handler.Rights.HasRights(context.Background(), roomID, playerID)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrInvalidHomeRoom
	}
	return nil
}

// player resolves the authenticated live player without allocations.
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

// applyRecord replaces live client settings after persistence succeeds.
func applyRecord(player *playerlive.Player, record Record) {
	player.SetClientSettings(record.VolumeSystem, record.VolumeFurniture, record.VolumeTrax, record.OldChat, record.CameraFollowBlocked, record.SafetyLocked)
}
