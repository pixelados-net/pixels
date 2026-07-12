// Package essential coordinates essential specialized furniture interactions.
package essential

import (
	"context"
	"errors"
	"math/rand/v2"
	"strconv"

	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	furnitureused "github.com/niflaot/pixels/internal/realm/furniture/events/used"
	furniturewalkedoff "github.com/niflaot/pixels/internal/realm/furniture/events/walkedoff"
	furniturewalkedon "github.com/niflaot/pixels/internal/realm/furniture/events/walkedon"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outdicevalue "github.com/niflaot/pixels/networking/outbound/room/furniture/dicevalue"
	outstate "github.com/niflaot/pixels/networking/outbound/room/furniture/state"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/i18n"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

var (
	// ErrNoRights reports a specialized interaction requiring room furniture rights.
	ErrNoRights = errors.New("specialized furniture interaction requires room rights")
)

// Source produces bounded random values.
type Source interface {
	// IntN returns a value in [0, limit).
	IntN(limit int) int
}

// Request describes one specialized furniture use.
type Request struct {
	// PlayerID identifies the acting player.
	PlayerID int64
	// Room stores the active room.
	Room *roomlive.Room
	// Item stores the targeted runtime furniture snapshot.
	Item worldfurniture.Item
}

// Service coordinates specialized furniture behavior.
type Service struct {
	// states persists final furniture states.
	states furnitureservice.StateUpdater
	// permissions resolves staff capabilities.
	permissions permissionservice.Checker
	// runtime stores active rooms.
	runtime *roomlive.Registry
	// connections sends room packets.
	connections *netconn.Registry
	// players stores online players.
	players *playerlive.Registry
	// events publishes interaction results.
	events bus.Publisher
	// translations resolves hotel-facing text.
	translations i18n.Translator
	// random provides deterministic bounded values.
	random Source
	// log records delayed interaction failures.
	log *zap.Logger
}

// defaultSource delegates to Go's concurrency-safe global random source.
type defaultSource struct{}

// IntN returns a bounded pseudo-random value.
func (defaultSource) IntN(limit int) int {
	return rand.IntN(limit)
}

// New creates the essential interaction service.
func New(states furnitureservice.StateUpdater, permissions permissionservice.Checker, runtime *roomlive.Registry, connections *netconn.Registry, players *playerlive.Registry, events *bus.Bus, translations i18n.Translator, log *zap.Logger) *Service {
	return &Service{
		states: states, permissions: permissions, runtime: runtime, connections: connections,
		players: players, events: events, translations: translations, random: defaultSource{}, log: log,
	}
}

// Use routes one click to its specialized behavior.
func (service *Service) Use(ctx context.Context, request Request) (bool, error) {
	if service == nil || request.Room == nil || request.PlayerID <= 0 || request.Item.ID <= 0 {
		return false, nil
	}
	switch request.Item.Definition.InteractionType {
	case "dice", "colorwheel", "random_state":
		return true, service.useRandom(ctx, request)
	case "pressureplate", "colorplate", "handitem_tile":
		return true, nil
	case "onewaygate", "switch", "switch_remote_control", "multiheight":
		return true, service.useTraversal(ctx, request)
	case "vendingmachine", "vendingmachine_no_sides", "handitem":
		return true, service.useHandItem(ctx, request)
	case "cannon":
		return true, service.useCannon(ctx, request)
	default:
		return false, nil
	}
}

// SetSource replaces randomness for deterministic tests.
func (service *Service) SetSource(source Source) {
	if source != nil {
		service.random = source
	}
}

// visual changes and broadcasts an ephemeral state.
func (service *Service) visual(ctx context.Context, active *roomlive.Room, itemID int64, value string) error {
	item, changed := active.SetFurnitureExtraData(itemID, value)
	if !changed {
		return nil
	}

	return service.broadcastState(ctx, active, item.ID, item.ExtraData)
}

// settle persists and broadcasts a final state.
func (service *Service) settle(ctx context.Context, active *roomlive.Room, itemID int64, expected string, value string, rebuild bool) error {
	_, err := service.states.UpdateState(ctx, furnitureservice.StateParams{
		ItemID: itemID, RoomID: active.ID(), Expected: expected, Next: value,
	})
	if err != nil {
		return err
	}
	item, err := active.UpdateFurnitureState(itemID, value, rebuild)
	if err != nil {
		return err
	}

	return service.broadcastState(ctx, active, item.ID, item.ExtraData)
}

// broadcastState sends one furniture state to every room occupant.
func (service *Service) broadcastState(ctx context.Context, active *roomlive.Room, itemID int64, value string) error {
	state, err := strconv.Atoi(value)
	if err != nil {
		state = 0
	}
	item, found := active.FurnitureItem(itemID)
	var packet codec.Packet
	if found && item.Definition.InteractionType == "dice" {
		packet, err = outdicevalue.Encode(itemID, state)
	} else {
		packet, err = outstate.Encode(itemID, state)
	}
	if err != nil {
		return err
	}

	return broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
}

// publishUsed emits one accepted furniture use.
func (service *Service) publishUsed(ctx context.Context, request Request) error {
	if service.events == nil {
		return nil
	}

	return service.events.Publish(ctx, bus.Event{Name: furnitureused.Name, Payload: furnitureused.Payload{
		PlayerID: request.PlayerID, ItemID: request.Item.ID, RoomID: request.Room.ID(),
	}})
}

// Register subscribes movement-driven essential interactions.
func Register(lifecycle fx.Lifecycle, subscriber bus.Subscriber, service *Service) error {
	on, err := subscriber.Subscribe(furniturewalkedon.Name, bus.PriorityHigh, service.walkedOn)
	if err != nil {
		return err
	}
	off, err := subscriber.Subscribe(furniturewalkedoff.Name, bus.PriorityHigh, service.walkedOff)
	if err != nil {
		on.Unsubscribe()
		return err
	}
	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
		on.Unsubscribe()
		off.Unsubscribe()
		return nil
	}})

	return nil
}
