package runtime

import (
	"context"
	"sync"
	"time"

	searchcmd "github.com/niflaot/pixels/internal/realm/navigator/browse/search"
	navservice "github.com/niflaot/pixels/internal/realm/navigator/core"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	roomoccupancy "github.com/niflaot/pixels/internal/realm/room/runtime/events/occupancychanged"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	netconn "github.com/niflaot/pixels/networking/connection"
	outsearch "github.com/niflaot/pixels/networking/outbound/navigator/browse/searchresult"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
)

const (
	// RoomCountDebounce is the navigator room counter refresh debounce window.
	RoomCountDebounce = 350 * time.Millisecond
)

// RoomCountBroadcaster refreshes visible navigator room counters.
type RoomCountBroadcaster struct {
	// mutex protects pending room refreshes.
	mutex sync.Mutex
	// players stores live players.
	players *playerlive.Registry
	// connections stores active protocol connections.
	connections *netconn.Registry
	// search builds navigator search results.
	search searchcmd.Handler
	// pending stores room ids that changed occupancy.
	pending map[int64]struct{}
	// timer flushes pending room ids after debounce.
	timer *time.Timer
	// delay stores the debounce window.
	delay time.Duration
}

// NewRoomCountBroadcaster creates a room count broadcaster.
func NewRoomCountBroadcaster(players *playerlive.Registry, connections *netconn.Registry, navigator navservice.Manager, rooms roomservice.Manager, runtime *roomlive.Registry) *RoomCountBroadcaster {
	return &RoomCountBroadcaster{
		players:     players,
		connections: connections,
		search:      searchcmd.Handler{Navigator: navigator, Rooms: rooms, Runtime: runtime},
		pending:     make(map[int64]struct{}),
		delay:       RoomCountDebounce,
	}
}

// RegisterRoomCounts subscribes navigator room count broadcasts.
func RegisterRoomCounts(lifecycle fx.Lifecycle, subscriber bus.Subscriber, broadcaster *RoomCountBroadcaster) error {
	subscription, err := subscriber.Subscribe(roomoccupancy.Name, bus.PriorityLow, broadcaster.Handle)
	if err != nil {
		return err
	}

	lifecycle.Append(fx.Hook{OnStop: func(context.Context) error {
		subscription.Unsubscribe()
		broadcaster.Close()
		return nil
	}})

	return nil
}

// Handle handles one room occupancy event.
func (broadcaster *RoomCountBroadcaster) Handle(ctx context.Context, event bus.Event) error {
	payload, ok := event.Payload.(roomoccupancy.Payload)
	if !ok || payload.RoomID <= 0 {
		return nil
	}

	broadcaster.queue(payload.RoomID)

	return nil
}

// Close stops pending broadcasts.
func (broadcaster *RoomCountBroadcaster) Close() {
	broadcaster.mutex.Lock()
	defer broadcaster.mutex.Unlock()

	if broadcaster.timer != nil {
		broadcaster.timer.Stop()
		broadcaster.timer = nil
	}
}

// Flush sends pending room counter refreshes.
func (broadcaster *RoomCountBroadcaster) Flush() {
	roomIDs := broadcaster.takePending()
	if len(roomIDs) == 0 || broadcaster.players == nil || broadcaster.connections == nil {
		return
	}

	for _, player := range broadcaster.players.Snapshot() {
		broadcaster.refreshPlayer(context.Background(), player, roomIDs)
	}
}

// queue stores a room id and schedules a flush.
func (broadcaster *RoomCountBroadcaster) queue(roomID int64) {
	broadcaster.mutex.Lock()
	defer broadcaster.mutex.Unlock()

	broadcaster.pending[roomID] = struct{}{}
	if broadcaster.timer == nil {
		broadcaster.timer = time.AfterFunc(broadcaster.delay, broadcaster.Flush)
	}
}

// refreshPlayer refreshes one affected navigator viewer.
func (broadcaster *RoomCountBroadcaster) refreshPlayer(ctx context.Context, player *playerlive.Player, roomIDs map[int64]struct{}) {
	viewer, found := player.Navigator()
	if !found || !viewer.HasAnyVisibleRoom(roomIDs) {
		return
	}

	search := viewer.LastSearch()
	if search.Code == "" {
		return
	}

	lists, _, err := broadcaster.search.Result(ctx, player.ID(), search.Code, search.Query)
	if err != nil {
		return
	}

	packet, err := outsearch.Encode(search.Code, search.Query, lists)
	if err != nil {
		return
	}

	peer := player.Peer()
	connection, found := broadcaster.connections.Get(peer.ConnectionKind(), peer.ConnectionID())
	if !found {
		return
	}

	viewer.SetSearch(search, searchcmd.VisibleRoomIDs(lists))
	_ = connection.Send(ctx, packet)
}

// takePending returns and clears pending room ids.
func (broadcaster *RoomCountBroadcaster) takePending() map[int64]struct{} {
	broadcaster.mutex.Lock()
	defer broadcaster.mutex.Unlock()

	roomIDs := broadcaster.pending
	broadcaster.pending = make(map[int64]struct{})
	broadcaster.timer = nil

	return roomIDs
}
