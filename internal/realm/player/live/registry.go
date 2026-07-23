package live

import "sync"

// Registry stores online live players.
type Registry struct {
	// mutex protects player storage.
	mutex sync.RWMutex

	// players stores live players by player id.
	players map[int64]*Player
}

// NewRegistry creates an empty live player registry.
func NewRegistry() *Registry {
	return &Registry{players: make(map[int64]*Player)}
}

// Add registers a live player.
func (registry *Registry) Add(player *Player) error {
	if player == nil || player.ID() <= 0 {
		return ErrInvalidPlayer
	}

	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	id := player.ID()
	if _, exists := registry.players[id]; exists {
		return ErrPlayerExists
	}

	registry.players[id] = player

	return nil
}

// Find returns a live player by id.
func (registry *Registry) Find(id int64) (*Player, bool) {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	player, found := registry.players[id]

	return player, found
}

// Remove deletes a live player by id.
func (registry *Registry) Remove(id int64) (*Player, bool) {
	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	player, found := registry.players[id]
	if !found {
		return nil, false
	}

	delete(registry.players, id)

	return player, true
}

// Count returns the number of online players.
func (registry *Registry) Count() int {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	return len(registry.players)
}

// Snapshot returns a stable copy of live players.
func (registry *Registry) Snapshot() []*Player {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	players := make([]*Player, 0, len(registry.players))
	for _, player := range registry.players {
		players = append(players, player)
	}

	return players
}

// NavigatorAudience returns players receiving navigator category counts.
func (registry *Registry) NavigatorAudience() []*Player {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	players := make([]*Player, 0, len(registry.players))
	for _, player := range registry.players {
		viewer, found := player.Navigator()
		if found && viewer.ReceivesCategoryCounts() {
			players = append(players, player)
		}
	}

	return players
}
