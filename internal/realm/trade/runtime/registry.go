package runtime

import "sync"

// Registry indexes live trades and staged furniture without hot-path allocation.
type Registry struct {
	// mutex protects all indexes.
	mutex sync.RWMutex
	// byPlayer maps both participants to one session.
	byPlayer map[int64]*Session
	// staged maps offered item ids to owning player ids.
	staged map[int64]int64
}

// NewRegistry creates an empty trade registry.
func NewRegistry() *Registry {
	return &Registry{byPlayer: make(map[int64]*Session), staged: make(map[int64]int64)}
}

// Start atomically creates one trade when both players are idle.
func (registry *Registry) Start(session *Session) bool {
	if session == nil || session.First.PlayerID <= 0 || session.Second.PlayerID <= 0 || session.First.PlayerID == session.Second.PlayerID {
		return false
	}
	registry.mutex.Lock()
	defer registry.mutex.Unlock()
	if registry.byPlayer[session.First.PlayerID] != nil || registry.byPlayer[session.Second.PlayerID] != nil {
		return false
	}
	registry.byPlayer[session.First.PlayerID] = session
	registry.byPlayer[session.Second.PlayerID] = session
	return true
}

// Find returns the player's live trade.
func (registry *Registry) Find(playerID int64) (*Session, bool) {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()
	session, found := registry.byPlayer[playerID]
	return session, found
}

// Stage atomically locks an item for a participant in its current session.
func (registry *Registry) Stage(playerID int64, itemID int64) bool {
	return registry.StageMany(playerID, []int64{itemID})
}

// StageMany atomically locks a complete offer mutation or none of it.
func (registry *Registry) StageMany(playerID int64, itemIDs []int64) bool {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()
	session := registry.byPlayer[playerID]
	if session == nil || len(itemIDs) == 0 {
		return false
	}
	for _, itemID := range itemIDs {
		if registry.staged[itemID] != 0 {
			return false
		}
	}
	if !session.AddItems(playerID, itemIDs) {
		return false
	}
	for _, itemID := range itemIDs {
		registry.staged[itemID] = playerID
	}
	return true
}

// Unstage removes one staged item from the participant's offer.
func (registry *Registry) Unstage(playerID int64, itemID int64) bool {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()
	session := registry.byPlayer[playerID]
	if session == nil || registry.staged[itemID] != playerID || !session.RemoveItem(playerID, itemID) {
		return false
	}
	delete(registry.staged, itemID)
	return true
}

// Contains reports whether an item is staged without allocating.
func (registry *Registry) Contains(itemID int64) bool {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()
	return registry.staged[itemID] != 0
}

// Close removes a player's whole session and every staged item.
func (registry *Registry) Close(playerID int64) (*Session, bool) {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()
	session := registry.byPlayer[playerID]
	if session == nil {
		return nil, false
	}
	if !session.TryCancel() {
		return nil, false
	}
	registry.remove(session)
	return session, true
}

// Complete removes one successfully settled session.
func (registry *Registry) Complete(playerID int64) (*Session, bool) {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()
	session := registry.byPlayer[playerID]
	if session == nil {
		return nil, false
	}
	session.CompleteSettlement()
	registry.remove(session)
	return session, true
}

// remove deletes one claimed session and every staged item.
func (registry *Registry) remove(session *Session) {
	first, second := session.Snapshot()
	delete(registry.byPlayer, first.PlayerID)
	delete(registry.byPlayer, second.PlayerID)
	for _, itemID := range first.Items {
		delete(registry.staged, itemID)
	}
	for _, itemID := range second.Items {
		delete(registry.staged, itemID)
	}
}

// ActiveCount returns live trade count.
func (registry *Registry) ActiveCount() int {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()
	return len(registry.byPlayer) / 2
}
