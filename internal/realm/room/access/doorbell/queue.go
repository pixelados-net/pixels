// Package doorbell stores players waiting for room entry approval.
package doorbell

import (
	"strings"
	"sync"
	"time"

	netconn "github.com/niflaot/pixels/networking/connection"
)

// ExpireReason describes why a waiting request was removed.
type ExpireReason uint8

const (
	// ExpiredTimeout means the request exceeded its hangout duration.
	ExpiredTimeout ExpireReason = iota + 1
	// ExpiredNoRightsHolder means nobody remains to answer the request.
	ExpiredNoRightsHolder
	// ExpiredRoomClosed means the active room was closed.
	ExpiredRoomClosed
)

// Entry describes one player waiting for room approval.
type Entry struct {
	// PlayerID identifies the waiting player.
	PlayerID int64
	// Username stores the protocol identity used by the response packet.
	Username string
	// Handler stores the waiting connection context.
	Handler netconn.Context
	// RequestedAt stores when the request was last refreshed.
	RequestedAt time.Time
}

// Valid reports whether an entry can receive a response.
func (entry Entry) Valid() bool {
	return entry.PlayerID > 0 && entry.Username != "" && entry.Handler.ConnectionID != "" && entry.Handler.ConnectionKind != "" && !entry.RequestedAt.IsZero()
}

// Expired describes one removed waiting request.
type Expired struct {
	Entry
	// Reason describes why the entry was removed.
	Reason ExpireReason
}

// Queue stores waiting players with lazy map allocation.
type Queue struct {
	// mutex protects queue entries.
	mutex sync.Mutex
	// entries stores requests by player id.
	entries map[int64]Entry
}

// Request creates or refreshes one waiting request.
func (queue *Queue) Request(entry Entry) bool {
	if !entry.Valid() {
		return false
	}
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	if queue.entries == nil {
		queue.entries = make(map[int64]Entry)
	}
	queue.entries[entry.PlayerID] = entry

	return true
}

// Resolve removes one request by case-insensitive username.
func (queue *Queue) Resolve(username string) (Entry, bool) {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	for playerID, entry := range queue.entries {
		if !strings.EqualFold(entry.Username, username) {
			continue
		}
		delete(queue.entries, playerID)
		queue.releaseEmpty()

		return entry, true
	}

	return Entry{}, false
}

// Sweep removes requests that reached their timeout.
func (queue *Queue) Sweep(now time.Time, timeout time.Duration) []Expired {
	if timeout <= 0 {
		return nil
	}
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	var expired []Expired
	for playerID, entry := range queue.entries {
		if entry.RequestedAt.Add(timeout).After(now) {
			continue
		}
		expired = append(expired, Expired{Entry: entry, Reason: ExpiredTimeout})
		delete(queue.entries, playerID)
	}
	queue.releaseEmpty()

	return expired
}

// Drain removes every waiting request with one reason.
func (queue *Queue) Drain(reason ExpireReason) []Expired {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()
	if len(queue.entries) == 0 {
		return nil
	}
	expired := make([]Expired, 0, len(queue.entries))
	for _, entry := range queue.entries {
		expired = append(expired, Expired{Entry: entry, Reason: reason})
	}
	queue.entries = nil

	return expired
}

// Len returns the number of waiting requests.
func (queue *Queue) Len() int {
	queue.mutex.Lock()
	defer queue.mutex.Unlock()

	return len(queue.entries)
}

// releaseEmpty releases the map after its final entry is removed.
func (queue *Queue) releaseEmpty() {
	if len(queue.entries) == 0 {
		queue.entries = nil
	}
}
