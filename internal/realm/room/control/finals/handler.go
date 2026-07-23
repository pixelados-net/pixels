// Package finals implements the remaining player-facing room control requests.
package finals

import (
	"context"

	"github.com/niflaot/pixels/internal/permission"
	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	furnituresession "github.com/niflaot/pixels/internal/realm/furniture/commands/session"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	ambassadoralerted "github.com/niflaot/pixels/internal/realm/room/control/events/ambassadoralerted"
	roomsettings "github.com/niflaot/pixels/internal/realm/room/control/settings"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inambassador "github.com/niflaot/pixels/networking/inbound/room/control/ambassador"
	indelete "github.com/niflaot/pixels/networking/inbound/room/control/delete"
	inobjectdata "github.com/niflaot/pixels/networking/inbound/room/control/objectdata"
	instaffpick "github.com/niflaot/pixels/networking/inbound/room/control/staffpick"
	outdesktop "github.com/niflaot/pixels/networking/outbound/session/desktop"
	"github.com/niflaot/pixels/pkg/bus"
)

// Handler coordinates room records, custom object data, and ambassador alerts.
type Handler struct {
	// Players stores live player state.
	Players *playerlive.Registry
	// Bindings maps authenticated connections to players.
	Bindings *binding.Registry
	// Rooms reads and deletes room records.
	Rooms roomservice.Manager
	// ConfigRooms updates room flags.
	ConfigRooms roomservice.ConfigManager
	// Settings authorizes owner room mutations.
	Settings *roomsettings.Authorizer
	// Furniture reads room furniture.
	Furniture furnitureservice.Manager
	// States persists furniture object data.
	States furnitureservice.StateUpdater
	// Runtime stores active rooms.
	Runtime *roomlive.Registry
	// Permissions resolves staff and ambassador capabilities.
	Permissions permissionservice.Checker
	// DeleteAny permits deleting another player's room.
	DeleteAny permission.Node
	// StaffPick permits changing official room selection.
	StaffPick permission.Node
	// Ambassador permits submitting ambassador alerts.
	Ambassador permission.Node
	// Connections stores active network connections.
	Connections *netconn.Registry
	// Events publishes moderation intake events.
	Events bus.Publisher
}

// Register adds the final room-control packet handlers.
func Register(registry *netconn.HandlerRegistry, handler Handler) {
	if registry == nil {
		return
	}
	_ = registry.Register(indelete.Header, handler.handleDelete)
	_ = registry.Register(instaffpick.Header, handler.handleStaffPick)
	_ = registry.Register(inambassador.Header, handler.handleAmbassador)
	_ = registry.Register(inobjectdata.Header, handler.handleObjectData)
}

// handleDelete soft-deletes an authorized room and closes its active runtime.
func (handler Handler) handleDelete(connection netconn.Context, packet codec.Packet) error {
	roomID, err := indelete.Decode(packet)
	if err != nil {
		return err
	}
	return handler.deleteRoom(context.Background(), connection, int64(roomID))
}

// deleteRoom applies owner or global deletion authorization.
func (handler Handler) deleteRoom(ctx context.Context, connection netconn.Context, roomID int64) error {
	player, room, err := handler.playerRoom(ctx, connection, roomID)
	if err != nil {
		return err
	}
	allowed, err := handler.Settings.CanManage(ctx, room, player.ID())
	if err != nil {
		return err
	}
	if !allowed {
		allowed, err = handler.has(ctx, player.ID(), handler.DeleteAny)
	}
	if err != nil || !allowed {
		return err
	}
	if err = handler.Rooms.SoftDelete(ctx, roomID); err != nil {
		return err
	}
	if active, found := handler.Runtime.Find(roomID); found {
		desktop, encodeErr := outdesktop.Encode()
		if encodeErr != nil {
			return encodeErr
		}
		_ = broadcast.RoomPacket(ctx, handler.Connections, active, desktop, 0)
		_, _, err = handler.Runtime.Close(ctx, roomID)
	}
	return err
}

// handleStaffPick changes official selection for an authorized staff player.
func (handler Handler) handleStaffPick(connection netconn.Context, packet codec.Packet) error {
	roomID, err := instaffpick.Decode(packet)
	if err != nil {
		return err
	}
	return handler.staffPick(context.Background(), connection, int64(roomID))
}

// staffPick sets the durable official-room flag.
func (handler Handler) staffPick(ctx context.Context, connection netconn.Context, roomID int64) error {
	player, room, err := handler.playerRoom(ctx, connection, roomID)
	if err != nil {
		return err
	}
	allowed, err := handler.has(ctx, player.ID(), handler.StaffPick)
	if err != nil || !allowed {
		return err
	}
	value := true
	_, err = handler.ConfigRooms.Update(ctx, room.ID, room.Version.Version, roomservice.UpdateParams{StaffPicked: &value, AllowReservedTags: true})
	return err
}

// handleAmbassador validates an in-room target and emits moderation intake.
func (handler Handler) handleAmbassador(connection netconn.Context, packet codec.Packet) error {
	targetID, err := inambassador.Decode(packet)
	if err != nil {
		return err
	}
	return handler.ambassador(context.Background(), connection, int64(targetID))
}

// ambassador publishes one authenticated room alert.
func (handler Handler) ambassador(ctx context.Context, connection netconn.Context, targetID int64) error {
	player, err := furnituresession.Player(connection, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	allowed, err := handler.has(ctx, player.ID(), handler.Ambassador)
	if err != nil || !allowed {
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
	if _, found = active.Occupant(targetID); !found || targetID == player.ID() {
		return nil
	}
	if handler.Events == nil {
		return nil
	}
	return handler.Events.Publish(ctx, bus.Event{Name: ambassadoralerted.Name, Payload: ambassadoralerted.Payload{ReporterPlayerID: player.ID(), ReportedPlayerID: targetID, RoomID: roomID}})
}

// handleObjectData persists bounded custom variables and projects them live.
func (handler Handler) handleObjectData(connection netconn.Context, packet codec.Packet) error {
	payload, err := inobjectdata.Decode(packet)
	if err != nil {
		return err
	}
	return handler.objectData(context.Background(), connection, payload)
}

// playerRoom resolves the authenticated player and one durable room.
func (handler Handler) playerRoom(ctx context.Context, connection netconn.Context, roomID int64) (*playerlive.Player, roommodel.Room, error) {
	player, err := furnituresession.Player(connection, handler.Bindings, handler.Players)
	if err != nil {
		return nil, roommodel.Room{}, err
	}
	room, found, err := handler.Rooms.FindByID(ctx, roomID)
	if err != nil || !found {
		return nil, roommodel.Room{}, err
	}
	return player, room, nil
}

// has resolves one concrete global permission.
func (handler Handler) has(ctx context.Context, playerID int64, node permission.Node) (bool, error) {
	if handler.Permissions == nil || !node.Concrete() {
		return false, nil
	}
	return handler.Permissions.HasPermission(ctx, playerID, node)
}
