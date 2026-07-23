package guide

import (
	"sort"
	"strings"
)

// Guardians returns oldest available guardian ids.
func (manager *Manager) Guardians(excludeID int64, limit int) []int64 {
	manager.mutex.RLock()
	defer manager.mutex.RUnlock()
	values := make([]Duty, 0)
	for id, duty := range manager.duty {
		_, busy := manager.byPlayer[id]
		if id != excludeID && duty.Guardian && !busy {
			values = append(values, duty)
		}
	}
	sort.Slice(values, func(i, j int) bool { return values[i].Since.Before(values[j].Since) })
	if len(values) > limit {
		values = values[:limit]
	}
	ids := make([]int64, len(values))
	for index := range values {
		ids[index] = values[index].PlayerID
	}
	return ids
}

// RemovePlayer clears duty and any active session on disconnect.
func (manager *Manager) RemovePlayer(playerID int64) (Session, bool) {
	manager.mutex.Lock()
	defer manager.mutex.Unlock()
	delete(manager.duty, playerID)
	delete(manager.completed, playerID)
	session := manager.sessionFor(playerID)
	if session == nil {
		return Session{}, false
	}
	value := clone(session)
	manager.remove(session)
	return value, true
}

// clean filters and bounds visible guide text.
func (manager *Manager) clean(value string) string {
	value = strings.Join(strings.Fields(value), " ")
	if manager.filter != nil {
		value, _ = manager.filter.Censor(value)
	}
	if len(value) > 500 {
		value = value[:500]
	}
	return value
}

// clone detaches mutable transcript storage.
func clone(session *Session) Session {
	value := *session
	value.Transcript = append([]Message(nil), session.Transcript...)
	return value
}
