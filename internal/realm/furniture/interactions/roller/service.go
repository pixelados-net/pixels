package roller

import (
	"context"
	"sync"

	rolledevent "github.com/niflaot/pixels/internal/realm/furniture/events/rolled"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	netconn "github.com/niflaot/pixels/networking/connection"
	outrolling "github.com/niflaot/pixels/networking/outbound/room/furniture/rolling"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/zap"
)

const persistenceQueueSize = 256

// Service coordinates room roller cycles, projections, and durable flushes.
type Service struct {
	// config stores normalized roller policy.
	config Config
	// furniture persists final furniture positions.
	furniture furnitureservice.Manager
	// connections broadcasts room animation packets.
	connections *netconn.Registry
	// events publishes completed roller facts and delayed walk hooks.
	events bus.Publisher
	// log records best-effort persistence failures.
	log *zap.Logger
	// persistence stores bounded asynchronous position updates.
	persistence chan persistence
	// stop ends the persistence worker.
	stop chan struct{}
	// done closes after the persistence worker exits.
	done chan struct{}
	// once protects worker lifecycle.
	once sync.Once
}

// New creates roller behavior.
func New(config Config, furniture furnitureservice.Manager, connections *netconn.Registry, events *bus.Bus, log *zap.Logger) *Service {
	service := &Service{
		config: config.Normalize(), furniture: furniture, connections: connections, log: log,
		persistence: make(chan persistence, persistenceQueueSize), stop: make(chan struct{}), done: make(chan struct{}),
	}
	if events != nil {
		service.events = events
	}
	return service
}

// Start begins the bounded persistence worker once.
func (service *Service) Start() {
	service.once.Do(func() { go service.runPersistence() })
}

// Stop flushes queued persistence updates and waits for the worker.
func (service *Service) Stop() {
	service.Start()
	close(service.stop)
	<-service.done
}

// NoRules reports whether roller placement restrictions are disabled.
func (service *Service) NoRules() bool {
	return service != nil && service.config.NoRules
}

// broadcast sends Nitro-compatible animations, heightmap updates, and events.
func (service *Service) broadcast(ctx context.Context, active *roomlive.Room, moved movedStep) error {
	units := make([]outrolling.Unit, 0, len(moved.units))
	for index, unit := range moved.units {
		units = append(units, outrolling.Unit{RoomIndex: unit.UnitID, FromZ: moved.unitSources[index].Position.Z.String(), ToZ: unit.Position.Z.String()})
	}
	items := make([]outrolling.Item, 0, len(moved.items))
	for index, item := range moved.items {
		items = append(items, outrolling.Item{ID: item.ID, FromZ: moved.itemSources[index].Z.String(), ToZ: item.Z.String()})
	}
	if service.connections != nil {
		if err := service.broadcastRolling(ctx, active, moved, units, items); err != nil {
			return err
		}
		if err := broadcast.RoomHeightMapUpdate(ctx, service.connections, active, []grid.Point{moved.step.roller.Point, moved.step.target}, 0); err != nil {
			return err
		}
	}
	return service.publishRolled(ctx, active.ID(), moved)
}

// broadcastRolling projects the one-unit legacy shape expected by Nitro.
func (service *Service) broadcastRolling(ctx context.Context, active *roomlive.Room, moved movedStep, units []outrolling.Unit, items []outrolling.Item) error {
	encode := func(packetItems []outrolling.Item, options ...outrolling.Option) error {
		packet, err := outrolling.Encode(
			int(moved.step.roller.Point.X), int(moved.step.roller.Point.Y),
			int(moved.step.target.X), int(moved.step.target.Y), packetItems, moved.step.roller.ID, options...,
		)
		if err != nil {
			return err
		}
		return broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
	}
	if len(units) == 0 {
		return encode(items)
	}
	for index, unit := range units {
		packetItems := items
		if index > 0 {
			packetItems = nil
		}
		if err := encode(packetItems, outrolling.WithUnit(unit)); err != nil {
			return err
		}
	}
	return nil
}

// publishRolled publishes one consolidated completed roller event.
func (service *Service) publishRolled(ctx context.Context, roomID int64, moved movedStep) error {
	if service.events == nil {
		return nil
	}
	itemIDs := make([]int64, 0, len(moved.items))
	for _, item := range moved.items {
		itemIDs = append(itemIDs, item.ID)
	}
	playerIDs := make([]int64, 0, len(moved.units))
	for _, unit := range moved.units {
		if unit.PlayerID > 0 {
			playerIDs = append(playerIDs, unit.PlayerID)
		}
	}
	return service.events.Publish(ctx, bus.Event{Name: rolledevent.Name, Payload: rolledevent.Payload{
		RoomID: roomID, RollerItemID: moved.step.roller.ID, ItemIDs: itemIDs, PlayerIDs: playerIDs,
		From: moved.step.roller.Point, To: moved.step.target,
	}})
}

// furniturePlacement maps a world item to persistence input.
func furniturePlacement(itemID int64, ownerID int64, roomID int64, x int, y int, z float64, rotation worldunit.Rotation) furnitureservice.MoveParams {
	return furnitureservice.MoveParams{ItemID: itemID, ActorPlayerID: ownerID, RoomID: roomID, Placement: furnituremodel.Placement{X: x, Y: y, Z: z, Rotation: furnituremodel.Rotation(rotation)}}
}
