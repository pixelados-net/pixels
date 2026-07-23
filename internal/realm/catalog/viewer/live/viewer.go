// Package live contains per-player catalog viewer state.
package live

import "sync"

const (
	// DefaultMode identifies the normal furniture catalog.
	DefaultMode = "NORMAL"
)

// Viewer stores one player's transient catalog UI state.
type Viewer struct {
	// mutex protects mutable viewer state.
	mutex sync.RWMutex
	// mode stores the active catalog mode.
	mode string
	// currentPageID identifies the last opened page.
	currentPageID int64
}

// NewViewer creates a catalog viewer in normal mode.
func NewViewer() *Viewer {
	return &Viewer{mode: DefaultMode}
}

// SetMode stores the active catalog mode.
func (viewer *Viewer) SetMode(mode string) {
	viewer.mutex.Lock()
	defer viewer.mutex.Unlock()
	if mode == "" {
		mode = DefaultMode
	}
	viewer.mode = mode
}

// Mode returns the active catalog mode.
func (viewer *Viewer) Mode() string {
	viewer.mutex.RLock()
	defer viewer.mutex.RUnlock()

	return viewer.mode
}

// SetPage stores the last opened catalog page.
func (viewer *Viewer) SetPage(pageID int64) {
	viewer.mutex.Lock()
	defer viewer.mutex.Unlock()
	viewer.currentPageID = pageID
}

// CurrentPage returns the last opened catalog page.
func (viewer *Viewer) CurrentPage() (int64, bool) {
	viewer.mutex.RLock()
	defer viewer.mutex.RUnlock()

	return viewer.currentPageID, viewer.currentPageID > 0
}
