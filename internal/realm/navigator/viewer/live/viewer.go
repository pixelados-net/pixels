package live

import (
	"sync"
	"time"
)

// LastSearch stores the viewer's last navigator query.
type LastSearch struct {
	// Code stores the search context or result code.
	Code string

	// Query stores the search query.
	Query string
}

// Viewer stores active navigator UI state for one player.
type Viewer struct {
	// mutex protects viewer state.
	mutex sync.RWMutex

	// initializedAt stores when the viewer initialized.
	initializedAt time.Time

	// lastSearch stores the last navigator search.
	lastSearch LastSearch

	// visibleRooms stores room ids in the current navigator result.
	visibleRooms map[int64]struct{}

	// categoryCounts reports whether category count updates are enabled.
	categoryCounts bool
}

// NewViewer creates a navigator viewer.
func NewViewer() *Viewer {
	return &Viewer{initializedAt: time.Now(), visibleRooms: make(map[int64]struct{}), categoryCounts: true}
}

// SetSearch replaces the viewer search state and visible rooms.
func (viewer *Viewer) SetSearch(search LastSearch, roomIDs []int64) {
	viewer.mutex.Lock()
	defer viewer.mutex.Unlock()

	viewer.lastSearch = search
	viewer.visibleRooms = roomSet(roomIDs)
}

// SetLastSearch replaces the viewer last search.
func (viewer *Viewer) SetLastSearch(search LastSearch) {
	viewer.SetSearch(search, viewer.VisibleRoomIDs())
}

// LastSearch returns the viewer last search.
func (viewer *Viewer) LastSearch() LastSearch {
	viewer.mutex.RLock()
	defer viewer.mutex.RUnlock()

	return viewer.lastSearch
}

// VisibleRoomIDs returns the current navigator result room ids.
func (viewer *Viewer) VisibleRoomIDs() []int64 {
	viewer.mutex.RLock()
	defer viewer.mutex.RUnlock()

	roomIDs := make([]int64, 0, len(viewer.visibleRooms))
	for roomID := range viewer.visibleRooms {
		roomIDs = append(roomIDs, roomID)
	}

	return roomIDs
}

// HasVisibleRoom reports whether the viewer currently sees a room.
func (viewer *Viewer) HasVisibleRoom(roomID int64) bool {
	viewer.mutex.RLock()
	defer viewer.mutex.RUnlock()

	_, found := viewer.visibleRooms[roomID]

	return found
}

// HasAnyVisibleRoom reports whether the viewer sees any room from the set.
func (viewer *Viewer) HasAnyVisibleRoom(roomIDs map[int64]struct{}) bool {
	viewer.mutex.RLock()
	defer viewer.mutex.RUnlock()

	for roomID := range roomIDs {
		if _, found := viewer.visibleRooms[roomID]; found {
			return true
		}
	}

	return false
}

// SetCategoryCounts enables or disables category count updates.
func (viewer *Viewer) SetCategoryCounts(enabled bool) {
	viewer.mutex.Lock()
	defer viewer.mutex.Unlock()

	viewer.categoryCounts = enabled
}

// ReceivesCategoryCounts reports whether the viewer receives category counts.
func (viewer *Viewer) ReceivesCategoryCounts() bool {
	viewer.mutex.RLock()
	defer viewer.mutex.RUnlock()

	return viewer.categoryCounts
}

// roomSet maps room ids into a lookup set.
func roomSet(roomIDs []int64) map[int64]struct{} {
	rooms := make(map[int64]struct{}, len(roomIDs))
	for _, roomID := range roomIDs {
		if roomID > 0 {
			rooms[roomID] = struct{}{}
		}
	}

	return rooms
}
