// Package game owns ephemeral room WIRED team and score state.
package game

import (
	"sync"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// State stores one room's ephemeral game state.
type State struct {
	// Running reports whether a game is active.
	Running bool
	// Teams maps player ids to teams one through four.
	Teams map[int64]int32
	// Scores stores individual player scores.
	Scores map[int64]int64
	// WiredScores stores score granted by WIRED and excluded from achievements.
	WiredScores map[int64]int64
	// TeamScores stores scores for teams one through four.
	TeamScores [5]int64
	// WiredUses counts each score effect per player in the current match.
	WiredUses map[int64]map[int64]int32
}

// Service stores game states by active room.
type Service struct {
	// mutex protects room state maps.
	mutex sync.RWMutex
	// rooms stores room game state.
	rooms map[int64]*State
	// roomWorlds resolves unit projections.
	roomWorlds *roomlive.Registry
	// connections broadcasts team effects.
	connections *netconn.Registry
	// highscores persists and resets durable boards.
	highscores record.HighscoreResetter
}

// New creates an empty WIRED game service.
func New() *Service { return &Service{rooms: make(map[int64]*State)} }

// NewProjected creates a game service with room team-effect projection.
func NewProjected(rooms *roomlive.Registry, connections *netconn.Registry, resetters ...record.HighscoreResetter) *Service {
	service := &Service{rooms: make(map[int64]*State), roomWorlds: rooms, connections: connections}
	if len(resetters) > 0 {
		service.highscores = resetters[0]
	}
	return service
}

// Start starts a new game, preserves chosen teams, and clears scores.
func (service *Service) Start(roomID int64) bool {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	state := service.stateLocked(roomID)
	if state.Running {
		return false
	}
	state.Running = true
	state.Scores = make(map[int64]int64)
	state.WiredScores = make(map[int64]int64)
	state.TeamScores = [5]int64{}
	state.WiredUses = make(map[int64]map[int64]int32)
	return true
}

// End ends an active game.
func (service *Service) End(roomID int64) bool {
	service.mutex.Lock()
	defer service.mutex.Unlock()
	state := service.rooms[roomID]
	if state == nil || !state.Running {
		return false
	}
	state.Running = false
	return true
}

// Reset removes one room's ephemeral game state.
func (service *Service) Reset(roomID int64) { service.Close(roomID) }

// Close releases one room's game state.
func (service *Service) Close(roomID int64) {
	service.mutex.Lock()
	delete(service.rooms, roomID)
	service.mutex.Unlock()
}

// Team reports one player's active team.
func (service *Service) Team(roomID int64, playerID int64) (int32, bool) {
	service.mutex.RLock()
	state := service.rooms[roomID]
	if state == nil {
		service.mutex.RUnlock()
		return 0, false
	}
	team, found := state.Teams[playerID]
	service.mutex.RUnlock()
	return team, found
}

// JoinTeam assigns one player before a match starts.
func (service *Service) JoinTeam(roomID int64, playerID int64, team int32) bool {
	if playerID <= 0 || team < 1 || team > 4 {
		return false
	}
	service.mutex.Lock()
	state := service.stateLocked(roomID)
	if state.Running {
		service.mutex.Unlock()
		return false
	}
	changed := state.Teams[playerID] != team
	state.Teams[playerID] = team
	service.mutex.Unlock()
	if changed {
		service.projectTeam(roomID, playerID, team)
	}
	return changed
}

// LeaveTeam removes one player from its current room team.
func (service *Service) LeaveTeam(roomID int64, playerID int64) bool {
	service.mutex.Lock()
	state := service.rooms[roomID]
	if state == nil || state.Running {
		service.mutex.Unlock()
		return false
	}
	if _, found := state.Teams[playerID]; !found {
		service.mutex.Unlock()
		return false
	}
	delete(state.Teams, playerID)
	service.mutex.Unlock()
	service.projectTeam(roomID, playerID, 0)
	return true
}

// RemovePlayer drops one departed participant and their per-player match state.
func (service *Service) RemovePlayer(roomID int64, playerID int64) bool {
	service.mutex.Lock()
	state := service.rooms[roomID]
	if state == nil {
		service.mutex.Unlock()
		return false
	}
	_, found := state.Teams[playerID]
	delete(state.Teams, playerID)
	delete(state.Scores, playerID)
	delete(state.WiredScores, playerID)
	delete(state.WiredUses, playerID)
	service.mutex.Unlock()
	if found {
		service.projectTeam(roomID, playerID, 0)
	}
	return found
}

// Snapshot returns a stable game-state copy.
func (service *Service) Snapshot(roomID int64) (State, bool) {
	service.mutex.RLock()
	state := service.rooms[roomID]
	if state == nil {
		service.mutex.RUnlock()
		return State{}, false
	}
	result := State{Running: state.Running, Teams: make(map[int64]int32, len(state.Teams)), Scores: make(map[int64]int64, len(state.Scores)), WiredScores: make(map[int64]int64, len(state.WiredScores)), TeamScores: state.TeamScores}
	for playerID, team := range state.Teams {
		result.Teams[playerID] = team
	}
	for playerID, score := range state.Scores {
		result.Scores[playerID] = score
	}
	for playerID, score := range state.WiredScores {
		result.WiredScores[playerID] = score
	}
	service.mutex.RUnlock()
	return result, true
}

// stateLocked returns a mutable room state under the caller lock.
func (service *Service) stateLocked(roomID int64) *State {
	state := service.rooms[roomID]
	if state == nil {
		state = &State{Teams: make(map[int64]int32), Scores: make(map[int64]int64), WiredScores: make(map[int64]int64), WiredUses: make(map[int64]map[int64]int32)}
		service.rooms[roomID] = state
	}
	return state
}
