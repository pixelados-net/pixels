package teleport

import (
	"context"
	"errors"
	"sync"
	"time"

	teleportpair "github.com/niflaot/pixels/internal/realm/furniture/interactions/teleport/pair"
	roomentry "github.com/niflaot/pixels/internal/realm/room/access/entry"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	roomitems "github.com/niflaot/pixels/internal/realm/room/world/items"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
)

var (
	// ErrInvalidUse reports malformed teleport use input.
	ErrInvalidUse = errors.New("invalid teleport use")
	// ErrNotTeleport reports use of another furniture behavior.
	ErrNotTeleport = errors.New("furniture item is not a teleport")
)

// Service coordinates paired teleport animation and transfer state.
type Service struct {
	// config stores operator teleport policy.
	config Config
	// pairs resolves durable target relationships.
	pairs *teleportpair.Service
	// runtime stores active rooms.
	runtime *roomlive.Registry
	// connections stores active transport connections.
	connections *netconn.Registry
	// entry grants optional destination access.
	entry *roomentry.Service
	// events publishes teleport lifecycle events.
	events bus.Publisher
	// rooms stores lazily allocated active transition shards.
	rooms sync.Map
	// pendingMutex protects cross-room destination handoff.
	pendingMutex sync.Mutex
	// pending stores cross-room destinations by player id.
	pending map[int64]pendingDestination
	// now supplies transition time.
	now func() time.Time
}

// NewService creates furniture teleport behavior.
func NewService(config Config, pairs *teleportpair.Service, runtime *roomlive.Registry, connections *netconn.Registry, entry *roomentry.Service, events bus.Publisher) *Service {
	return &Service{config: config, pairs: pairs, runtime: runtime, connections: connections, entry: entry, events: events, now: time.Now}
}

// Start begins one teleport transition when the source and pair are valid.
func (service *Service) Start(ctx context.Context, request StartRequest) error {
	if request.PlayerID <= 0 || request.Room == nil || request.ItemID <= 0 {
		return ErrInvalidUse
	}
	source, found := request.Room.FurnitureItem(request.ItemID)
	if !found || !isTeleport(source) {
		return ErrNotTeleport
	}
	state := service.roomState(request.Room.ID())
	state.mutex.Lock()
	if _, active := state.transits[request.PlayerID]; active {
		state.mutex.Unlock()
		return nil
	}
	state.transits[request.PlayerID] = Transit{
		PlayerID: request.PlayerID, Source: source, SourceRoomID: request.Room.ID(), Phase: PhaseResolving,
	}
	state.mutex.Unlock()
	targetRecord, targetDefinition, found, err := service.pairs.FindTarget(ctx, request.ItemID)
	if err != nil || !found {
		service.removeTransit(request.Room.ID(), request.PlayerID)
		return err
	}
	if targetRecord.RoomID == nil || targetRecord.X == nil || targetRecord.Y == nil || targetRecord.Z == nil {
		service.removeTransit(request.Room.ID(), request.PlayerID)
		return ErrInvalidUse
	}
	targetRoomID := *targetRecord.RoomID
	worldDefinition, err := roomitems.ToWorldDefinition(targetDefinition)
	if err != nil {
		service.removeTransit(request.Room.ID(), request.PlayerID)
		return err
	}
	targetPoint, valid := grid.NewPoint(*targetRecord.X, *targetRecord.Y)
	if !valid {
		service.removeTransit(request.Room.ID(), request.PlayerID)
		return ErrInvalidUse
	}
	target := worldfurniture.Item{
		ID: targetRecord.ID, OwnerPlayerID: targetRecord.OwnerPlayerID,
		Definition: worldDefinition,
		Point:      targetPoint,
		Z:          roomitems.RoundHeight(*targetRecord.Z), Rotation: worldunit.Rotation(targetRecord.Rotation),
		ExtraData: targetRecord.ExtraData,
	}
	if activeTarget, active := service.runtime.Find(targetRoomID); active {
		if item, itemFound := activeTarget.FurnitureItem(target.ID); itemFound {
			target = item
		}
	}
	state.mutex.Lock()
	transit := Transit{PlayerID: request.PlayerID, Source: source, Target: target, SourceRoomID: request.Room.ID(), TargetRoomID: targetRoomID, Phase: PhaseApproach}
	state.transits[request.PlayerID] = transit
	state.mutex.Unlock()
	if err := service.startApproach(ctx, request.Room, transit); err != nil {
		service.removeTransit(request.Room.ID(), request.PlayerID)
		return err
	}

	return service.publishStarted(ctx, request.Room.ID(), transit)
}

// removeTransit removes one reserved or active transition.
func (service *Service) removeTransit(roomID int64, playerID int64) {
	loaded, found := service.rooms.Load(roomID)
	if !found {
		return
	}
	state := loaded.(*roomState)
	state.mutex.Lock()
	delete(state.transits, playerID)
	empty := len(state.transits) == 0
	state.mutex.Unlock()
	if empty {
		service.rooms.Delete(roomID)
	}
}

// roomState returns a lazily allocated transition shard.
func (service *Service) roomState(roomID int64) *roomState {
	created := &roomState{transits: make(map[int64]Transit)}
	actual, _ := service.rooms.LoadOrStore(roomID, created)

	return actual.(*roomState)
}

// isTeleport reports whether an item uses paired teleport behavior.
func isTeleport(item worldfurniture.Item) bool {
	return item.Definition.InteractionType == "teleport" || item.Definition.InteractionType == "teleport_tile"
}
