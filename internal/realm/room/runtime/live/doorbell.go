package live

import (
	"time"

	roomdoorbell "github.com/niflaot/pixels/internal/realm/room/access/doorbell"
)

// RequestDoorbell queues a player when an authorized responder is present.
func (room *Room) RequestDoorbell(entry roomdoorbell.Entry, approverPresent bool) bool {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	if room.closed || !approverPresent {
		return false
	}
	queue := room.doorbell.Load()
	if queue == nil {
		queue = &roomdoorbell.Queue{}
		room.doorbell.Store(queue)
	}

	return queue.Request(entry)
}

// ResolveDoorbell removes one waiting request by username.
func (room *Room) ResolveDoorbell(username string) (roomdoorbell.Entry, bool) {
	room.mutex.RLock()
	defer room.mutex.RUnlock()
	queue := room.doorbell.Load()
	if queue == nil {
		return roomdoorbell.Entry{}, false
	}

	return queue.Resolve(username)
}

// SweepDoorbell removes timed-out waiting requests.
func (room *Room) SweepDoorbell(now time.Time, timeout time.Duration) []roomdoorbell.Expired {
	queue := room.doorbell.Load()
	if queue == nil {
		return nil
	}
	room.mutex.Lock()
	defer room.mutex.Unlock()
	queue = room.doorbell.Load()
	if queue == nil {
		return nil
	}
	expired := queue.Sweep(now, timeout)
	if queue.Len() == 0 {
		room.doorbell.CompareAndSwap(queue, nil)
	}

	return expired
}

// DrainDoorbellWithoutApprover removes requests when no authorized responder remains.
func (room *Room) DrainDoorbellWithoutApprover(approverPresent bool) []roomdoorbell.Expired {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	queue := room.doorbell.Load()
	if queue == nil || approverPresent {
		return nil
	}
	room.doorbell.CompareAndSwap(queue, nil)
	expired := queue.Drain(roomdoorbell.ExpiredNoRightsHolder)

	return expired
}

// DoorbellLen returns the number of waiting requests.
func (room *Room) DoorbellLen() int {
	queue := room.doorbell.Load()
	if queue == nil {
		return 0
	}

	return queue.Len()
}

// OwnerPresent reports whether the room owner is currently inside.
func (room *Room) OwnerPresent() bool {
	room.mutex.RLock()
	defer room.mutex.RUnlock()

	return room.ownerPresentLocked()
}

// ownerPresentLocked reports owner presence while a room lock is held.
func (room *Room) ownerPresentLocked() bool {
	_, found := room.occupants[room.snapshot.OwnerPlayerID]

	return found
}

// SetUnitIdle replaces one unit's AFK projection.
func (room *Room) SetUnitIdle(playerID int64, idle bool) (UnitSnapshot, bool) {
	return room.SetUnitIdleAt(playerID, idle, time.Now())
}

// SetUnitIdleAt replaces one unit's AFK projection at one deterministic instant.
func (room *Room) SetUnitIdleAt(playerID int64, idle bool, at time.Time) (UnitSnapshot, bool) {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	if room.world == nil {
		return UnitSnapshot{}, false
	}
	return room.world.SetUnitIdleAt(playerID, idle, at)
}

// SetUnitPosture changes one unit's free-standing posture.
func (room *Room) SetUnitPosture(playerID int64, sitting bool) (UnitSnapshot, bool) {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	if room.world == nil {
		return UnitSnapshot{}, false
	}
	return room.world.SetUnitPosture(playerID, sitting)
}

// SetUnitDance changes one unit's persistent dance state.
func (room *Room) SetUnitDance(playerID int64, danceID int32) (UnitSnapshot, bool) {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	if room.world == nil {
		return UnitSnapshot{}, false
	}
	return room.world.SetUnitDance(playerID, danceID)
}

// SetUnitEffect replaces one unit's selected avatar effect.
func (room *Room) SetUnitEffect(playerID int64, effectID int32) (UnitSnapshot, bool) {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	if room.world == nil {
		return UnitSnapshot{}, false
	}
	return room.world.SetUnitEffect(playerID, effectID)
}

// RemainingMute returns one active mute duration without persistence I/O.
func (room *Room) RemainingMute(playerID int64, now time.Time) (time.Duration, bool) {
	room.mutex.RLock()
	endsAt, found := room.muted[playerID]
	room.mutex.RUnlock()
	if !found || !endsAt.After(now) {
		return 0, false
	}

	return endsAt.Sub(now), true
}
