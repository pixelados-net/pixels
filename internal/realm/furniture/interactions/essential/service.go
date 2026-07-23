// Package essential coordinates essential specialized furniture interactions.
package essential

import (
	"context"
	"errors"
	"math/rand/v2"
	"strconv"
	"time"

	permissionservice "github.com/niflaot/pixels/internal/permission/service"
	furnitureused "github.com/niflaot/pixels/internal/realm/furniture/events/used"
	furniturewalkedoff "github.com/niflaot/pixels/internal/realm/furniture/events/walkedoff"
	furniturewalkedon "github.com/niflaot/pixels/internal/realm/furniture/events/walkedon"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	playereffect "github.com/niflaot/pixels/internal/realm/player/effect"
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

const effectFurnitureDurationSeconds int32 = 86400

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
	// State stores the client interaction surface parameter.
	State int32
	// Target stores the acting protocol connection when one initiated the use.
	Target netconn.Context
}

// External handles one specialized interaction owned by another realm.
type External interface {
	// UseFurniture handles matching furniture and reports whether it claimed the use.
	UseFurniture(context.Context, Request) (bool, error)
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
	// effects grants and enables furniture effects.
	effects playereffect.Manager
	// translations resolves hotel-facing text.
	translations i18n.Translator
	// random provides deterministic bounded values.
	random Source
	// log records delayed interaction failures.
	log *zap.Logger
	// external stores startup-registered cross-realm interaction handlers.
	external []External
}

// defaultSource delegates to Go's concurrency-safe global random source.
type defaultSource struct{}

// IntN returns a bounded pseudo-random value.
func (defaultSource) IntN(limit int) int {
	return rand.IntN(limit)
}

// New creates the essential interaction service.
func New(states furnitureservice.StateUpdater, permissions permissionservice.Checker, runtime *roomlive.Registry, connections *netconn.Registry, players *playerlive.Registry, events *bus.Bus, translations i18n.Translator, log *zap.Logger) *Service {
	service := &Service{
		states: states, permissions: permissions, runtime: runtime, connections: connections,
		players: players, translations: translations, random: defaultSource{}, log: log,
	}
	if events != nil {
		service.events = events
	}
	return service
}

// NewWithEffects creates the production essential service with effect grants enabled.
func NewWithEffects(states furnitureservice.StateUpdater, permissions permissionservice.Checker, runtime *roomlive.Registry, connections *netconn.Registry, players *playerlive.Registry, events *bus.Bus, effects playereffect.Manager, translations i18n.Translator, log *zap.Logger) *Service {
	service := New(states, permissions, runtime, connections, players, events, translations, log)
	service.effects = effects
	return service
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
	case "effect_giver":
		return true, service.useEffectGiver(ctx, request)
	default:
		for _, handler := range service.external {
			if handled, err := handler.UseFurniture(ctx, request); handled || err != nil {
				return handled, err
			}
		}
		return false, nil
	}
}

// AddExternal registers one cross-realm specialized interaction.
func (service *Service) AddExternal(handler External) {
	if service != nil && handler != nil {
		service.external = append(service.external, handler)
	}
}

// useEffectGiver grants and immediately enables one random configured effect.
func (service *Service) useEffectGiver(ctx context.Context, request Request) error {
	pool := request.Item.Definition.EffectPool
	if service.effects == nil || len(pool) == 0 {
		return nil
	}
	unit, found := request.Room.Unit(request.PlayerID)
	if !found || !onItem(unit.Position.Point, request.Item) && !adjacentToItem(unit.Position.Point, request.Item) {
		return nil
	}
	effectID := pool[service.random.IntN(len(pool))]
	if _, err := service.effects.GrantEnabled(ctx, request.PlayerID, effectID, effectFurnitureDurationSeconds, playereffect.SourceEffectGiver); err != nil {
		return err
	}
	if err := service.publishUsed(ctx, request); err != nil {
		return err
	}
	if request.Item.Definition.InteractionModesCount <= 1 {
		return nil
	}
	if err := service.visual(ctx, request.Room, request.Item.ID, "1"); err != nil {
		return err
	}
	async := context.WithoutCancel(ctx)
	request.Room.ScheduleReplacing(scheduledKey(request.Item.ID, 6), 500*time.Millisecond, func(time.Time) {
		_ = service.visual(async, request.Room, request.Item.ID, "0")
	})
	return nil
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
