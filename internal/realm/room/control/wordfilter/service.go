// Package wordfilter manages room-specific chat word filters.
package wordfilter

import (
	"context"
	"errors"
	"strings"
	"sync"
	"unicode/utf8"

	roomsettings "github.com/niflaot/pixels/internal/realm/room/control/settings"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	"github.com/niflaot/pixels/pkg/textfilter"
)

const (
	// MaxWordLength stores the maximum persisted filter word length.
	MaxWordLength = 32
	// MaxCachedRooms bounds process-local immutable filter snapshots.
	MaxCachedRooms = 1024
)

var (
	// ErrInvalidWord reports malformed filter input.
	ErrInvalidWord = errors.New("invalid room filter word")
	// ErrRoomNotFound reports a missing room.
	ErrRoomNotFound = errors.New("room word filter room not found")
)

// Manager reads and mutates room-specific filtered words.
type Manager interface {
	// List lists room filter words.
	List(context.Context, int64) ([]string, error)
	// Add adds a room filter word after authorization.
	Add(context.Context, int64, int64, string) error
	// Remove removes a room filter word after authorization.
	Remove(context.Context, int64, int64, string) error
	// Contains reports whether text contains a filtered whole word.
	Contains(context.Context, int64, string) (bool, error)
	// Censor replaces filtered whole words while preserving Unicode length.
	Censor(context.Context, int64, string) (string, bool, error)
}

// RoomFinder reads room identity and ownership.
type RoomFinder interface {
	// FindByID finds one active room record.
	FindByID(context.Context, int64) (roommodel.Room, bool, error)
}

// Snapshot stores one room's immutable filter generation.
type Snapshot struct {
	// words stores normalized persistent entries.
	words []string
	// matcher stores the compiled Aho-Corasick automaton.
	matcher *textfilter.Matcher
}

// Store persists room-specific filtered words.
type Store interface {
	// List lists normalized words for a room.
	List(context.Context, int64) ([]string, error)
	// Add inserts a normalized room word.
	Add(context.Context, int64, string) error
	// Remove deletes a normalized room word.
	Remove(context.Context, int64, string) error
}

// Service manages persistent filters and immutable cached snapshots.
type Service struct {
	// store persists room filter words.
	store Store
	// rooms reads room metadata.
	rooms RoomFinder
	// authorize resolves settings capability.
	authorize *roomsettings.Authorizer
	// cache stores immutable word slices by room id.
	cache map[int64]*Snapshot
	// cacheMutex protects immutable filter snapshot references.
	cacheMutex sync.RWMutex
}

// New creates a room word filter service.
func New(store Store, rooms RoomFinder, authorize *roomsettings.Authorizer) *Service {
	return &Service{store: store, rooms: rooms, authorize: authorize, cache: make(map[int64]*Snapshot)}
}

// List lists room filter words and returns a caller-owned slice.
func (service *Service) List(ctx context.Context, roomID int64) ([]string, error) {
	snapshot, err := service.snapshotFor(ctx, roomID)
	if err != nil {
		return nil, err
	}

	return append([]string(nil), snapshot.words...), nil
}

// Add adds a room filter word after authorization.
func (service *Service) Add(ctx context.Context, roomID int64, actorID int64, word string) error {
	word, err := service.validateMutation(ctx, roomID, actorID, word)
	if err != nil {
		return err
	}
	if err = service.store.Add(ctx, roomID, word); err != nil {
		return err
	}
	service.invalidate(roomID)

	return nil
}

// Remove removes a room filter word after authorization.
func (service *Service) Remove(ctx context.Context, roomID int64, actorID int64, word string) error {
	word, err := service.validateMutation(ctx, roomID, actorID, word)
	if err != nil {
		return err
	}
	if err = service.store.Remove(ctx, roomID, word); err != nil {
		return err
	}
	service.invalidate(roomID)

	return nil
}

// Contains reports whether text contains a normalized filtered pattern.
func (service *Service) Contains(ctx context.Context, roomID int64, text string) (bool, error) {
	snapshot, err := service.snapshotFor(ctx, roomID)
	if err != nil || len(snapshot.words) == 0 {
		return false, err
	}
	return snapshot.matcher.Contains(text), nil
}

// Censor replaces filtered patterns while preserving separators.
func (service *Service) Censor(ctx context.Context, roomID int64, text string) (string, bool, error) {
	snapshot, err := service.snapshotFor(ctx, roomID)
	if err != nil || len(snapshot.words) == 0 {
		return text, false, err
	}

	result, changed := snapshot.matcher.Censor(text)

	return result, changed, nil
}

// snapshotFor returns one immutable cached filter generation.
func (service *Service) snapshotFor(ctx context.Context, roomID int64) (*Snapshot, error) {
	service.cacheMutex.RLock()
	cached, found := service.cache[roomID]
	service.cacheMutex.RUnlock()
	if found {
		return cached, nil
	}
	words, err := service.store.List(ctx, roomID)
	if err != nil {
		return nil, err
	}
	service.cacheMutex.Lock()
	if len(service.cache) >= MaxCachedRooms {
		for cachedRoomID := range service.cache {
			delete(service.cache, cachedRoomID)
			break
		}
	}
	snapshot := &Snapshot{words: words, matcher: textfilter.Compile(words)}
	service.cache[roomID] = snapshot
	service.cacheMutex.Unlock()

	return snapshot, nil
}

// invalidate removes one cached room filter snapshot.
func (service *Service) invalidate(roomID int64) {
	service.cacheMutex.Lock()
	delete(service.cache, roomID)
	service.cacheMutex.Unlock()
}

// validateMutation normalizes and authorizes one filter mutation.
func (service *Service) validateMutation(ctx context.Context, roomID int64, actorID int64, word string) (string, error) {
	word = strings.ToLower(strings.TrimSpace(word))
	if roomID <= 0 || actorID <= 0 || word == "" || utf8.RuneCountInString(word) > MaxWordLength {
		return "", ErrInvalidWord
	}
	room, found, err := service.rooms.FindByID(ctx, roomID)
	if err != nil {
		return "", err
	}
	if !found {
		return "", ErrRoomNotFound
	}
	if err = service.authorize.Authorize(ctx, room, actorID); err != nil {
		return "", err
	}

	return word, nil
}

// managerAssertion verifies Service implements Manager.
var managerAssertion Manager = (*Service)(nil)
