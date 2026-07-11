package entry

import (
	"sync"
	"time"
)

// TrustStore stores short-lived server-issued room entry bypasses.
type TrustStore struct {
	// mutex protects trusted entries.
	mutex sync.Mutex
	// entries stores expiration by player and room.
	entries map[trustKey]time.Time
}

// trustKey identifies one trusted room entry.
type trustKey struct {
	// playerID identifies the target player.
	playerID int64
	// roomID identifies the target room.
	roomID int64
}

// Grant stores a trusted entry until its expiration.
func (store *TrustStore) Grant(playerID int64, roomID int64, expiresAt time.Time) bool {
	if playerID <= 0 || roomID <= 0 || expiresAt.IsZero() {
		return false
	}
	store.mutex.Lock()
	defer store.mutex.Unlock()
	if store.entries == nil {
		store.entries = make(map[trustKey]time.Time)
	}
	store.entries[trustKey{playerID: playerID, roomID: roomID}] = expiresAt

	return true
}

// Consume removes and validates one trusted entry.
func (store *TrustStore) Consume(playerID int64, roomID int64, now time.Time) bool {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	if store.entries == nil {
		return false
	}
	key := trustKey{playerID: playerID, roomID: roomID}
	expiresAt, found := store.entries[key]
	if !found {
		return false
	}
	delete(store.entries, key)
	if len(store.entries) == 0 {
		store.entries = nil
	}

	return expiresAt.After(now)
}

// Len returns the number of pending trusted entries.
func (store *TrustStore) Len() int {
	store.mutex.Lock()
	defer store.mutex.Unlock()

	return len(store.entries)
}
