package live

// ProfileViewer stores one player's current public-profile UI state.
type ProfileViewer struct {
	// playerID identifies the profile currently opened by the viewer.
	playerID int64
}

// ViewProfile replaces the currently observed public profile.
func (player *Player) ViewProfile(playerID int64) bool {
	if playerID <= 0 {
		return false
	}
	player.mutex.Lock()
	player.profileViewer.playerID = playerID
	player.mutex.Unlock()
	return true
}

// ViewedProfile returns the currently observed public profile.
func (player *Player) ViewedProfile() (int64, bool) {
	player.mutex.RLock()
	playerID := player.profileViewer.playerID
	player.mutex.RUnlock()
	return playerID, playerID > 0
}

// CloseProfile clears the current public-profile UI state.
func (player *Player) CloseProfile() (int64, bool) {
	player.mutex.Lock()
	playerID := player.profileViewer.playerID
	player.profileViewer.playerID = 0
	player.mutex.Unlock()
	return playerID, playerID > 0
}

// ReplaceIgnored atomically replaces the player's ignored-user projection.
func (player *Player) ReplaceIgnored(playerIDs []int64) {
	ignored := make(map[int64]struct{}, len(playerIDs))
	for _, playerID := range playerIDs {
		if playerID > 0 && playerID != player.ID() {
			ignored[playerID] = struct{}{}
		}
	}
	player.mutex.Lock()
	player.ignored = ignored
	player.mutex.Unlock()
}

// Ignore adds one player to the live ignored-user projection.
func (player *Player) Ignore(playerID int64) bool {
	if playerID <= 0 || playerID == player.ID() {
		return false
	}
	player.mutex.Lock()
	defer player.mutex.Unlock()
	if _, found := player.ignored[playerID]; found {
		return false
	}
	player.ignored[playerID] = struct{}{}
	return true
}

// Unignore removes one player from the live ignored-user projection.
func (player *Player) Unignore(playerID int64) bool {
	player.mutex.Lock()
	defer player.mutex.Unlock()
	if _, found := player.ignored[playerID]; !found {
		return false
	}
	delete(player.ignored, playerID)
	return true
}

// IsIgnoring reports whether this player hides communication from another player.
func (player *Player) IsIgnoring(playerID int64) bool {
	player.mutex.RLock()
	_, found := player.ignored[playerID]
	player.mutex.RUnlock()
	return found
}
