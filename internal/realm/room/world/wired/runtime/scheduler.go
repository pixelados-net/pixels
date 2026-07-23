package runtime

import (
	"time"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
)

// RoomScheduler queues delayed effects on the existing active-room task queue.
type RoomScheduler struct {
	// rooms resolves active room lifecycle owners.
	rooms *roomlive.Registry
}

// NewRoomScheduler creates a scheduler backed by active rooms.
func NewRoomScheduler(rooms *roomlive.Registry) *RoomScheduler { return &RoomScheduler{rooms: rooms} }

// Schedule queues work only when the room remains active.
func (scheduler *RoomScheduler) Schedule(roomID int64, _ uint64, delay time.Duration, run func(time.Time)) bool {
	if scheduler == nil || scheduler.rooms == nil || run == nil {
		return false
	}
	room, found := scheduler.rooms.Find(roomID)
	if !found {
		return false
	}
	room.Schedule(delay, run)
	return true
}

// releaseDelayed removes one discarded generation's outstanding task gauge.
func (engine *Engine) releaseDelayed(loaded *state) {
	loaded.mutex.Lock()
	delayed := loaded.delayed
	loaded.delayed = 0
	loaded.mutex.Unlock()
	if delayed > 0 {
		engine.metrics.delayedTasks.Add(-int64(delayed))
	}
}
