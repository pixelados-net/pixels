package live

import "github.com/niflaot/pixels/internal/realm/room/world/grid"

// HasUnitAt reports whether any room unit currently owns a tile.
func (room *Room) HasUnitAt(point grid.Point) bool {
	room.mutex.RLock()
	defer room.mutex.RUnlock()
	return room.world != nil && room.world.HasUnitAt(point)
}

// UpdateOccupantProfile replaces visible avatar fields for one active player.
func (room *Room) UpdateOccupantProfile(playerID int64, figure string, gender string, motto string) bool {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	occupant, found := room.occupants[playerID]
	if !found {
		return false
	}
	occupant.Figure = figure
	occupant.Gender = gender
	occupant.Motto = motto
	room.occupants[playerID] = occupant
	return true
}

// UpdateOccupantName replaces one active player's visible username.
func (room *Room) UpdateOccupantName(playerID int64, username string) bool {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	occupant, found := room.occupants[playerID]
	if !found {
		return false
	}
	occupant.Username = username
	room.occupants[playerID] = occupant
	return true
}

// UpdateOccupantGroup replaces one active player's favorite social-group projection.
func (room *Room) UpdateOccupantGroup(playerID int64, groupID int64, status int32, name string) bool {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	occupant, found := room.occupants[playerID]
	if !found {
		return false
	}
	occupant.GroupID = groupID
	occupant.GroupStatus = status
	occupant.GroupName = name
	room.occupants[playerID] = occupant
	return true
}

// UpdateOccupantAchievementScore replaces one active player's visible progression score.
func (room *Room) UpdateOccupantAchievementScore(playerID int64, score int32) bool {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	occupant, found := room.occupants[playerID]
	if !found {
		return false
	}
	occupant.AchievementScore = score
	room.occupants[playerID] = occupant
	return true
}
