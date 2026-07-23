package runtime

import (
	"context"
	"sort"
	"sync"
	"time"

	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomoccupancy "github.com/niflaot/pixels/internal/realm/room/runtime/events/occupancychanged"
	netconn "github.com/niflaot/pixels/networking/connection"
	outcounts "github.com/niflaot/pixels/networking/outbound/navigator/browse/categorycounts"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
)

const (
	// CategoryCountDebounce is the initial navigator count debounce window.
	CategoryCountDebounce = 500 * time.Millisecond
)

// CategoryCountBroadcaster sends navigator category count updates.
type CategoryCountBroadcaster struct {
	// mutex protects pending category counts.
	mutex sync.Mutex
	// players stores live players.
	players *playerlive.Registry
	// connections stores active protocol connections.
	connections *netconn.Registry
	// pending stores pending count entries by category id.
	pending map[int32]outcounts.Entry
	// rooms stores last active occupancy by room id.
	rooms map[int64]roomoccupancy.Payload
	// timer flushes pending entries after debounce.
	timer *time.Timer
	// delay stores the debounce window.
	delay time.Duration
}

// NewCategoryCountBroadcaster creates a category count broadcaster.
func NewCategoryCountBroadcaster(players *playerlive.Registry, connections *netconn.Registry) *CategoryCountBroadcaster {
	return &CategoryCountBroadcaster{
		players:     players,
		connections: connections,
		pending:     make(map[int32]outcounts.Entry),
		rooms:       make(map[int64]roomoccupancy.Payload),
		delay:       CategoryCountDebounce,
	}
}

// RegisterCategoryCounts subscribes navigator category count broadcasts.
func RegisterCategoryCounts(lifecycle fx.Lifecycle, subscriber bus.Subscriber, broadcaster *CategoryCountBroadcaster) error {
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
func (broadcaster *CategoryCountBroadcaster) Handle(ctx context.Context, event bus.Event) error {
	payload, ok := event.Payload.(roomoccupancy.Payload)
	if !ok || payload.CategoryID == nil {
		return nil
	}

	broadcaster.queue(payload)

	return nil
}

// Close stops pending broadcasts.
func (broadcaster *CategoryCountBroadcaster) Close() {
	broadcaster.mutex.Lock()
	defer broadcaster.mutex.Unlock()

	if broadcaster.timer != nil {
		broadcaster.timer.Stop()
		broadcaster.timer = nil
	}
}

// Snapshot returns current category count entries.
func (broadcaster *CategoryCountBroadcaster) Snapshot() []outcounts.Entry {
	broadcaster.mutex.Lock()
	defer broadcaster.mutex.Unlock()

	entries := make(map[int32]outcounts.Entry)
	for _, payload := range broadcaster.rooms {
		if payload.CategoryID == nil {
			continue
		}

		categoryID := int32(*payload.CategoryID)
		entry := entries[categoryID]
		entry.CategoryID = categoryID
		entry.CurrentVisitorCount += int32(payload.Count)
		entry.MaxVisitorCount += int32(payload.MaxUsers)
		entries[categoryID] = entry
	}

	return sortedEntries(entries)
}

// queue stores room occupancy and schedules a flush.
func (broadcaster *CategoryCountBroadcaster) queue(payload roomoccupancy.Payload) {
	broadcaster.mutex.Lock()
	defer broadcaster.mutex.Unlock()

	categoryID := int32(*payload.CategoryID)
	if payload.Count <= 0 {
		delete(broadcaster.rooms, payload.RoomID)
	} else {
		broadcaster.rooms[payload.RoomID] = payload
	}
	broadcaster.pending[categoryID] = broadcaster.categoryEntryLocked(categoryID)
	if broadcaster.timer == nil {
		broadcaster.timer = time.AfterFunc(broadcaster.delay, broadcaster.Flush)
	}
}

// categoryEntryLocked aggregates active rooms while locked.
func (broadcaster *CategoryCountBroadcaster) categoryEntryLocked(categoryID int32) outcounts.Entry {
	entry := outcounts.Entry{CategoryID: categoryID}
	for _, payload := range broadcaster.rooms {
		if payload.CategoryID != nil && int32(*payload.CategoryID) == categoryID {
			entry.CurrentVisitorCount += int32(payload.Count)
			entry.MaxVisitorCount += int32(payload.MaxUsers)
		}
	}

	return entry
}

// Flush sends pending category count entries.
func (broadcaster *CategoryCountBroadcaster) Flush() {
	entries := broadcaster.takePending()
	if len(entries) == 0 {
		return
	}

	packet, err := outcounts.Encode(entries)
	if err != nil {
		return
	}

	for _, player := range broadcaster.players.NavigatorAudience() {
		peer := player.Peer()
		connection, found := broadcaster.connections.Get(peer.ConnectionKind(), peer.ConnectionID())
		if found {
			_ = connection.Send(context.Background(), packet)
		}
	}
}

// takePending returns and clears pending count entries.
func (broadcaster *CategoryCountBroadcaster) takePending() []outcounts.Entry {
	broadcaster.mutex.Lock()
	defer broadcaster.mutex.Unlock()

	entries := sortedEntries(broadcaster.pending)
	broadcaster.pending = make(map[int32]outcounts.Entry)
	broadcaster.timer = nil

	return entries
}

// sortedEntries maps entries to stable protocol order.
func sortedEntries(entries map[int32]outcounts.Entry) []outcounts.Entry {
	results := make([]outcounts.Entry, 0, len(entries))
	for _, entry := range entries {
		results = append(results, entry)
	}
	sort.Slice(results, func(left int, right int) bool {
		return results[left].CategoryID < results[right].CategoryID
	})

	return results
}
