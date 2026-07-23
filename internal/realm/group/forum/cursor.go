package forum

import (
	"sync"
	"time"

	groupconfig "github.com/niflaot/pixels/internal/realm/group/config"
)

// Cursor stores the authorized forum context behind header-only CFH packets.
type Cursor struct {
	// PlayerID identifies the authenticated viewer.
	PlayerID int64
	// GroupID identifies the opened forum.
	GroupID int64
	// ThreadID identifies the opened thread.
	ThreadID int64
	// MessageID identifies the latest explicitly opened message when present.
	MessageID int64
	// ViewedAt stores successful read time.
	ViewedAt time.Time
}

// Cursors stores bounded ephemeral context per authenticated connection.
type Cursors struct {
	// ttl controls stale cursor rejection.
	ttl time.Duration
	// values stores current context by connection key.
	values sync.Map
}

// NewCursors creates ephemeral forum cursor storage.
func NewCursors(config groupconfig.Config) *Cursors { return &Cursors{ttl: config.ForumCursorTTL} }

// Set replaces one connection cursor.
func (cursors *Cursors) Set(connectionID string, cursor Cursor) {
	cursor.ViewedAt = time.Now()
	cursors.values.Store(connectionID, cursor)
}

// Get returns one fresh cursor or removes a stale value.
func (cursors *Cursors) Get(connectionID string) (Cursor, bool) {
	value, found := cursors.values.Load(connectionID)
	if !found {
		return Cursor{}, false
	}
	cursor := value.(Cursor)
	if time.Since(cursor.ViewedAt) > cursors.ttl {
		cursors.values.Delete(connectionID)
		return Cursor{}, false
	}
	return cursor, true
}

// Close removes one connection cursor.
func (cursors *Cursors) Close(connectionID string) { cursors.values.Delete(connectionID) }

// Viewers returns fresh distinct player viewers for one forum and prunes stale entries.
func (cursors *Cursors) Viewers(groupID int64) []int64 {
	now := time.Now()
	seen := make(map[int64]struct{})
	players := make([]int64, 0)
	cursors.values.Range(func(key, value any) bool {
		cursor := value.(Cursor)
		if now.Sub(cursor.ViewedAt) > cursors.ttl {
			cursors.values.Delete(key)
			return true
		}
		if cursor.GroupID != groupID || cursor.PlayerID <= 0 {
			return true
		}
		if _, found := seen[cursor.PlayerID]; found {
			return true
		}
		seen[cursor.PlayerID] = struct{}{}
		players = append(players, cursor.PlayerID)
		return true
	})
	return players
}
