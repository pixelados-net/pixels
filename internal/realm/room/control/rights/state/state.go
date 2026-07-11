// Package state stores the active projection of room build rights.
package state

import "sync"

// State stores one room owner's active build-right holders.
type State struct {
	// mutex protects the active rights projection.
	mutex sync.RWMutex
	// ownerID identifies the room owner.
	ownerID int64
	// holders stores explicit right holders by player id.
	holders map[int64]struct{}
}

// New creates an empty active rights projection.
func New(ownerID int64) *State {
	return &State{ownerID: ownerID}
}

// ReplaceRights replaces every explicit right holder.
func (state *State) ReplaceRights(playerIDs []int64) {
	holders := make(map[int64]struct{}, len(playerIDs))
	for _, playerID := range playerIDs {
		if playerID > 0 {
			holders[playerID] = struct{}{}
		}
	}
	state.mutex.Lock()
	state.holders = holders
	state.mutex.Unlock()
}

// GrantRights adds one explicit right holder.
func (state *State) GrantRights(playerID int64) {
	if playerID <= 0 {
		return
	}
	state.mutex.Lock()
	if state.holders == nil {
		state.holders = make(map[int64]struct{})
	}
	state.holders[playerID] = struct{}{}
	state.mutex.Unlock()
}

// RevokeRights removes one explicit right holder.
func (state *State) RevokeRights(playerID int64) {
	state.mutex.Lock()
	delete(state.holders, playerID)
	state.mutex.Unlock()
}

// HasRights reports whether a player owns or holds room rights.
func (state *State) HasRights(playerID int64) bool {
	if playerID <= 0 {
		return false
	}
	state.mutex.RLock()
	_, found := state.holders[playerID]
	owner := state.ownerID == playerID
	state.mutex.RUnlock()

	return owner || found
}

// Clear removes every explicit right holder.
func (state *State) Clear() {
	state.mutex.Lock()
	state.holders = nil
	state.mutex.Unlock()
}
