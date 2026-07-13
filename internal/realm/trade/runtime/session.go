// Package runtime owns live direct-trade sessions and staged item indexes.
package runtime

import "sync"

// Participant stores one side of a live trade.
type Participant struct {
	// PlayerID identifies the durable player.
	PlayerID int64
	// UnitID identifies the room-local unit.
	UnitID int64
	// Username stores the visible name.
	Username string
	// IP stores the connection address used by optional audit logging.
	IP string
	// Items stores offered furniture ids in insertion order.
	Items []int64
	// Accepted reports first-phase acceptance.
	Accepted bool
	// Confirmed reports final confirmation.
	Confirmed bool
}

// Session stores one synchronized live trade.
type Session struct {
	// mutex protects both participant states.
	mutex sync.RWMutex
	// RoomID identifies the shared active room.
	RoomID int64
	// First stores the initiating participant.
	First Participant
	// Second stores the invited participant.
	Second Participant
	// settling reports that one goroutine owns final persistence.
	settling bool
	// closed reports that cancellation or completion owns the session.
	closed bool
}

// Snapshot returns stable participant copies.
func (session *Session) Snapshot() (Participant, Participant) {
	session.mutex.RLock()
	defer session.mutex.RUnlock()
	first, second := session.First, session.Second
	first.Items = append([]int64(nil), first.Items...)
	second.Items = append([]int64(nil), second.Items...)
	return first, second
}

// Participant returns one participant snapshot.
func (session *Session) Participant(playerID int64) (Participant, bool) {
	first, second := session.Snapshot()
	if first.PlayerID == playerID {
		return first, true
	}
	if second.PlayerID == playerID {
		return second, true
	}
	return Participant{}, false
}

// SetIP stores one participant's normalized audit address.
func (session *Session) SetIP(playerID int64, ip string) bool {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	participant := session.mutable(playerID)
	if participant == nil {
		return false
	}
	participant.IP = ip
	return true
}

// Other returns the opposite participant snapshot.
func (session *Session) Other(playerID int64) (Participant, bool) {
	first, second := session.Snapshot()
	if first.PlayerID == playerID {
		return second, true
	}
	if second.PlayerID == playerID {
		return first, true
	}
	return Participant{}, false
}

// AddItem appends an item and resets both agreement phases.
func (session *Session) AddItem(playerID int64, itemID int64) bool {
	return session.AddItems(playerID, []int64{itemID})
}

// AddItems appends a whole validated offer mutation and resets both agreement phases.
func (session *Session) AddItems(playerID int64, itemIDs []int64) bool {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	participant := session.mutable(playerID)
	if participant == nil || session.closed || session.settling || len(itemIDs) == 0 {
		return false
	}
	for index, itemID := range itemIDs {
		for _, existing := range participant.Items {
			if existing == itemID {
				return false
			}
		}
		for prior := 0; prior < index; prior++ {
			if itemIDs[prior] == itemID {
				return false
			}
		}
	}
	participant.Items = append(participant.Items, itemIDs...)
	session.reset()
	return true
}

// RemoveItem removes an offered item and resets both agreement phases.
func (session *Session) RemoveItem(playerID int64, itemID int64) bool {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	participant := session.mutable(playerID)
	if participant == nil || session.closed || session.settling {
		return false
	}
	for index, existing := range participant.Items {
		if existing == itemID {
			participant.Items = append(participant.Items[:index], participant.Items[index+1:]...)
			session.reset()
			return true
		}
	}
	return false
}

// SetAccepted updates first-phase acceptance and returns whether both accepted.
func (session *Session) SetAccepted(playerID int64, accepted bool) (bool, bool) {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	participant := session.mutable(playerID)
	if participant == nil || session.closed || session.settling {
		return false, false
	}
	participant.Accepted = accepted
	participant.Confirmed = false
	if !accepted {
		if other := session.otherMutable(playerID); other != nil {
			other.Confirmed = false
		}
	}
	return session.First.Accepted && session.Second.Accepted, true
}

// Confirm updates final confirmation and returns whether settlement may run.
func (session *Session) Confirm(playerID int64) (bool, bool) {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	participant := session.mutable(playerID)
	if participant == nil || session.closed || !session.First.Accepted || !session.Second.Accepted {
		return false, false
	}
	participant.Confirmed = true
	ready := session.First.Confirmed && session.Second.Confirmed
	if !ready || session.settling {
		return false, true
	}
	session.settling = true
	return true, true
}

// TryCancel atomically claims cancellation unless settlement is running.
func (session *Session) TryCancel() bool {
	session.mutex.Lock()
	defer session.mutex.Unlock()
	if session.closed || session.settling {
		return false
	}
	session.closed = true
	return true
}

// CompleteSettlement marks a successfully settled session closed.
func (session *Session) CompleteSettlement() {
	session.mutex.Lock()
	session.closed = true
	session.settling = false
	session.mutex.Unlock()
}

// FailSettlement releases final confirmation after a transactional failure.
func (session *Session) FailSettlement() {
	session.mutex.Lock()
	session.First.Confirmed = false
	session.Second.Confirmed = false
	session.settling = false
	session.mutex.Unlock()
}

// mutable returns one participant under the caller's write lock.
func (session *Session) mutable(playerID int64) *Participant {
	if session.First.PlayerID == playerID {
		return &session.First
	}
	if session.Second.PlayerID == playerID {
		return &session.Second
	}
	return nil
}

// otherMutable returns the opposite participant under the caller's write lock.
func (session *Session) otherMutable(playerID int64) *Participant {
	if session.First.PlayerID == playerID {
		return &session.Second
	}
	if session.Second.PlayerID == playerID {
		return &session.First
	}
	return nil
}

// reset clears both agreement phases under the caller's write lock.
func (session *Session) reset() {
	session.First.Accepted = false
	session.Second.Accepted = false
	session.First.Confirmed = false
	session.Second.Confirmed = false
	session.settling = false
}
