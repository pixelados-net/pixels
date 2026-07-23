package binding

import (
	"sync"
	"time"
)

// Registry stores live player connection bindings.
type Registry struct {
	// mutex protects binding maps.
	mutex sync.RWMutex
	// byPlayer stores bindings by player id.
	byPlayer map[int64]Binding
	// byConnection stores player ids by connection key.
	byConnection map[ConnectionKey]int64
}

// NewRegistry creates an empty binding registry.
func NewRegistry() *Registry {
	return &Registry{
		byPlayer:     make(map[int64]Binding),
		byConnection: make(map[ConnectionKey]int64),
	}
}

// Add registers a player connection binding.
func (registry *Registry) Add(binding Binding) error {
	if !binding.Valid() {
		return ErrInvalidBinding
	}

	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	key := ConnectionKey{Kind: binding.ConnectionKind, ID: binding.ConnectionID}
	if _, exists := registry.byPlayer[binding.PlayerID]; exists {
		return ErrBindingExists
	}
	if _, exists := registry.byConnection[key]; exists {
		return ErrBindingExists
	}

	binding = binding.WithTime(time.Now())
	registry.byPlayer[binding.PlayerID] = binding
	registry.byConnection[key] = binding.PlayerID

	return nil
}

// FindByPlayer returns a binding by player id.
func (registry *Registry) FindByPlayer(playerID int64) (Binding, bool) {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	binding, found := registry.byPlayer[playerID]

	return binding, found
}

// FindByConnection returns a binding by connection key.
func (registry *Registry) FindByConnection(key ConnectionKey) (Binding, bool) {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	playerID, found := registry.byConnection[key]
	if !found {
		return Binding{}, false
	}

	binding, found := registry.byPlayer[playerID]

	return binding, found
}

// RemoveByPlayer removes a binding by player id.
func (registry *Registry) RemoveByPlayer(playerID int64) (Binding, bool) {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	binding, found := registry.byPlayer[playerID]
	if !found {
		return Binding{}, false
	}

	delete(registry.byPlayer, playerID)
	delete(registry.byConnection, ConnectionKey{Kind: binding.ConnectionKind, ID: binding.ConnectionID})

	return binding, true
}

// RemoveByConnection removes a binding by connection key.
func (registry *Registry) RemoveByConnection(key ConnectionKey) (Binding, bool) {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	playerID, found := registry.byConnection[key]
	if !found {
		return Binding{}, false
	}

	binding := registry.byPlayer[playerID]
	delete(registry.byConnection, key)
	delete(registry.byPlayer, playerID)

	return binding, true
}

// Count returns the number of active bindings.
func (registry *Registry) Count() int {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	return len(registry.byPlayer)
}

// Snapshot returns a stable copy of current bindings.
func (registry *Registry) Snapshot() []Binding {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	bindings := make([]Binding, 0, len(registry.byPlayer))
	for _, binding := range registry.byPlayer {
		bindings = append(bindings, binding)
	}

	return bindings
}
