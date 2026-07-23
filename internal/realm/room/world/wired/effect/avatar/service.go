// Package avatar executes player-facing WIRED room effects.
package avatar

import (
	"context"
	"strconv"
	"strings"
	"unicode/utf8"

	playerachievement "github.com/niflaot/pixels/internal/realm/player/achievement"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roommoderation "github.com/niflaot/pixels/internal/realm/room/control/moderation"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roomprojection "github.com/niflaot/pixels/internal/realm/room/runtime/projection"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	chatwhisper "github.com/niflaot/pixels/networking/outbound/chat/whisper"
	roomfx "github.com/niflaot/pixels/networking/outbound/room/entities/effect"
	roomhand "github.com/niflaot/pixels/networking/outbound/room/entities/handitem"
	sessionalert "github.com/niflaot/pixels/networking/outbound/session/alert"
)

// Service executes player effects against active-room primitives.
type Service struct {
	// rooms resolves active rooms.
	rooms *roomlive.Registry
	// players resolves the current authenticated population for safe alert tokens.
	players *playerlive.Registry
	// connections sends player and room projections.
	connections *netconn.Registry
	// moderation applies durable room-only kick and mute actions.
	moderation *roommoderation.Service
	// achievements grants durable respect.
	achievements *playerachievement.Service
}

// New creates a player-facing WIRED effect service.
func New(rooms *roomlive.Registry, players *playerlive.Registry, connections *netconn.Registry, moderation *roommoderation.Service, achievements *playerachievement.Service) *Service {
	return &Service{rooms: rooms, players: players, connections: connections, moderation: moderation, achievements: achievements}
}

// ExecuteAvatar executes one validated player-facing operation.
func (service *Service) ExecuteAvatar(ctx context.Context, operation effect.AvatarOperation, node *configuration.Node, event trigger.Event) (effect.Result, error) {
	active, found := service.rooms.Find(event.RoomID)
	if !found {
		return effect.Result{Status: effect.Skipped}, nil
	}
	if operation == effect.TeleportAvatar {
		entityKey := event.ActorID
		if entityKey == 0 {
			entityKey = event.PlayerID
		}
		unit, exists := active.UnitMotion(entityKey)
		if !exists {
			return effect.Result{Status: effect.Skipped}, nil
		}
		return service.teleport(active, node, entityKey, unit)
	}
	if operation == effect.ShowMessage && event.PlayerID <= 0 {
		return service.sendRoomMessage(ctx, active, node.Parameters.Text, event)
	}
	if event.PlayerID <= 0 {
		return effect.Result{Status: effect.Skipped}, nil
	}
	unit, found := active.Unit(event.PlayerID)
	if !found {
		return effect.Result{Status: effect.Skipped}, nil
	}
	switch operation {
	case effect.ShowMessage:
		packet, err := chatwhisper.Encode(int32(unit.UnitID), node.Parameters.Text, 0, 0, int32(utf8.RuneCountInString(node.Parameters.Text)))
		return service.sendPlayer(ctx, active, event.PlayerID, packet, err)
	case effect.AlertAvatar:
		online := 0
		if service.players != nil {
			online = service.players.Count()
		}
		packet, err := sessionalert.Encode(expand(node.Parameters.Text, event, online, service.rooms.Count()))
		return service.sendPlayer(ctx, active, event.PlayerID, packet, err)
	case effect.KickAvatar:
		if service.moderation == nil {
			return effect.Result{Status: effect.Blocked}, nil
		}
		return effect.Result{Status: effect.Applied}, service.moderation.SystemKick(ctx, event.RoomID, event.PlayerID)
	case effect.MuteAvatar:
		if service.moderation == nil || len(node.Parameters.Values) == 0 {
			return effect.Result{Status: effect.Blocked}, nil
		}
		return effect.Result{Status: effect.Applied}, service.moderation.SystemMute(ctx, event.RoomID, event.PlayerID, node.Parameters.Values[0])
	case effect.GiveHanditem:
		return service.handitem(ctx, active, event.PlayerID, unit.UnitID, node.Parameters.Number)
	case effect.GiveEffect:
		return service.giveEffect(ctx, active, event.PlayerID, unit.UnitID, node.Parameters.Number)
	case effect.GiveRespect:
		if service.achievements == nil {
			return effect.Result{Status: effect.Blocked}, nil
		}
		key := "wired:" + strconv.FormatInt(node.ItemID, 10) + ":" + strconv.FormatUint(event.ID, 10) + ":" + strconv.FormatInt(event.PlayerID, 10)
		granted, err := service.achievements.GrantRespect(ctx, event.PlayerID, node.Parameters.Number, key)
		if err != nil {
			return effect.Result{Status: effect.Blocked}, err
		}
		if !granted {
			return effect.Result{Status: effect.Skipped}, nil
		}
		return resultApplied(), nil
	default:
		return effect.Result{Status: effect.Blocked}, nil
	}
}

// sendRoomMessage delivers actorless WIRED messages privately to every current player.
func (service *Service) sendRoomMessage(ctx context.Context, active *roomlive.Room, message string, event trigger.Event) (effect.Result, error) {
	if service.connections == nil || message == "" {
		return effect.Result{Status: effect.Skipped}, nil
	}
	delivered := false
	for _, occupant := range active.Occupants() {
		unit, found := active.Unit(occupant.PlayerID)
		if !found {
			continue
		}
		recipient := event
		recipient.PlayerID = occupant.PlayerID
		recipient.ActorID = occupant.PlayerID
		recipient.Username = occupant.Username
		resolved := expand(message, recipient, service.onlineCount(), service.rooms.Count())
		packet, err := chatwhisper.Encode(int32(unit.UnitID), resolved, 0, 0, int32(utf8.RuneCountInString(resolved)))
		if err != nil {
			return effect.Result{Status: effect.Blocked}, err
		}
		connection, found := service.connections.Get(occupant.ConnectionKind, occupant.ConnectionID)
		if !found {
			continue
		}
		_ = connection.Send(ctx, packet)
		delivered = true
	}
	if !delivered {
		return effect.Result{Status: effect.Skipped}, nil
	}
	return resultApplied(), nil
}

// onlineCount returns the current authenticated population for message tokens.
func (service *Service) onlineCount() int {
	if service.players == nil {
		return 0
	}
	return service.players.Count()
}

// sendPlayer sends one effect packet only to the actor.
func (service *Service) sendPlayer(ctx context.Context, active *roomlive.Room, playerID int64, packet codec.Packet, encodeErr error) (effect.Result, error) {
	if encodeErr != nil {
		return effect.Result{Status: effect.Blocked}, encodeErr
	}
	occupant, found := active.Occupant(playerID)
	if !found || service.connections == nil {
		return effect.Result{Status: effect.Skipped}, nil
	}
	connection, found := service.connections.Get(occupant.ConnectionKind, occupant.ConnectionID)
	if !found {
		return effect.Result{Status: effect.Skipped}, nil
	}
	return resultApplied(), connection.Send(ctx, packet)
}

// teleport repositions the actor onto the first valid selected furniture target.
func (service *Service) teleport(active *roomlive.Room, node *configuration.Node, entityKey int64, unit roomlive.UnitSnapshot) (effect.Result, error) {
	for _, target := range node.Targets {
		item, found := active.FurnitureItem(target.ItemID)
		if !found {
			continue
		}
		moved, err := active.TeleportUnit(entityKey, item.Point, unit.BodyRotation, false, roomlive.TeleportNear)
		if err != nil {
			continue
		}
		return resultApplied(), broadcast.RoomUnitStatuses(context.Background(), service.connections, active, []roomlive.UnitSnapshot{moved}, 0)
	}
	return effect.Result{Status: effect.Skipped}, nil
}

// handitem updates and broadcasts the actor's room hand item.
func (service *Service) handitem(ctx context.Context, active *roomlive.Room, playerID int64, unitID int64, itemID int32) (effect.Result, error) {
	if itemID < 0 || itemID > 9999 {
		return effect.Result{Status: effect.Blocked}, nil
	}
	if _, err := active.SetHandItem(playerID, itemID); err != nil {
		return effect.Result{Status: effect.Skipped}, nil
	}
	packet, err := roomhand.Encode(unitID, itemID)
	if err != nil {
		return effect.Result{Status: effect.Blocked}, err
	}
	return resultApplied(), broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
}

// giveEffect updates and broadcasts the actor's room-scoped effect.
func (service *Service) giveEffect(ctx context.Context, active *roomlive.Room, playerID int64, unitID int64, effectID int32) (effect.Result, error) {
	if effectID < 0 || effectID > 10000 {
		return effect.Result{Status: effect.Blocked}, nil
	}
	unit, found := active.SetUnitEffect(playerID, effectID)
	if !found {
		return effect.Result{Status: effect.Skipped}, nil
	}
	packet, err := roomfx.Encode(unitID, roomprojection.EffectID(unit), 0)
	if err != nil {
		return effect.Result{Status: effect.Blocked}, err
	}
	return resultApplied(), broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
}

// expand replaces the closed compatibility token catalog.
func expand(message string, event trigger.Event, online int, rooms int) string {
	message = strings.ReplaceAll(message, "%username%", event.Username)
	message = strings.ReplaceAll(message, "%online%", strconv.Itoa(online))
	message = strings.ReplaceAll(message, "%roomsloaded%", strconv.Itoa(rooms))
	return message
}

// resultApplied returns a successful effect result.
func resultApplied() effect.Result { return effect.Result{Status: effect.Applied} }
