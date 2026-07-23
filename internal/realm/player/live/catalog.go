package live

import catalogviewer "github.com/niflaot/pixels/internal/realm/catalog/viewer/live"

// OpenCatalog creates or returns the player's catalog viewer.
func (player *Player) OpenCatalog() *catalogviewer.Viewer {
	player.mutex.Lock()
	defer player.mutex.Unlock()
	if player.catalog == nil {
		player.catalog = catalogviewer.NewViewer()
	}

	return player.catalog
}

// Catalog returns the player's catalog viewer.
func (player *Player) Catalog() (*catalogviewer.Viewer, bool) {
	player.mutex.RLock()
	defer player.mutex.RUnlock()

	return player.catalog, player.catalog != nil
}

// CloseCatalog removes the player's catalog viewer.
func (player *Player) CloseCatalog() (*catalogviewer.Viewer, bool) {
	player.mutex.Lock()
	defer player.mutex.Unlock()
	if player.catalog == nil {
		return nil, false
	}
	viewer := player.catalog
	player.catalog = nil

	return viewer, true
}
