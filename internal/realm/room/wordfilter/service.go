// Package wordfilter manages room-specific chat word filters.
package wordfilter

import (
	"context"
	"errors"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
	roomsettings "github.com/niflaot/pixels/internal/realm/room/settings"
	wordrepo "github.com/niflaot/pixels/internal/realm/room/wordfilter/repository"
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
}

// RoomFinder reads room identity and ownership.
type RoomFinder interface {
	// FindByID finds one active room record.
	FindByID(context.Context, int64) (roommodel.Room, bool, error)
}

// Service manages persistent filters and immutable cached snapshots.
type Service struct {
	// store persists room filter words.
	store wordrepo.Store
	// rooms reads room metadata.
	rooms RoomFinder
	// authorize resolves settings capability.
	authorize *roomsettings.Authorizer
	// cache stores immutable word slices by room id.
	cache map[int64][]string
	// cacheMutex protects immutable filter snapshot references.
	cacheMutex sync.RWMutex
}

// New creates a room word filter service.
func New(store wordrepo.Store, rooms RoomFinder, authorize *roomsettings.Authorizer) *Service {
	return &Service{store: store, rooms: rooms, authorize: authorize, cache: make(map[int64][]string)}
}

// List lists room filter words and returns a caller-owned slice.
func (service *Service) List(ctx context.Context, roomID int64) ([]string, error) {
	words, err := service.words(ctx, roomID)
	if err != nil {
		return nil, err
	}

	return append([]string(nil), words...), nil
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

// Contains reports whether text contains a filtered whole word.
func (service *Service) Contains(ctx context.Context, roomID int64, text string) (bool, error) {
	words, err := service.words(ctx, roomID)
	if err != nil || len(words) == 0 {
		return false, err
	}
	for start := 0; start < len(text); {
		var end int
		start, end = nextToken(text, start)
		if start == end {
			break
		}
		token := text[start:end]
		for _, word := range words {
			if strings.EqualFold(token, word) {
				return true, nil
			}
		}
		start = end
	}

	return false, nil
}

// words returns one immutable cached filter snapshot.
func (service *Service) words(ctx context.Context, roomID int64) ([]string, error) {
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
	service.cache[roomID] = words
	service.cacheMutex.Unlock()

	return words, nil
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

// nextToken returns the next letter-or-number token bounds without allocation.
func nextToken(text string, offset int) (int, int) {
	start := offset
	for start < len(text) {
		r, size := utf8.DecodeRuneInString(text[start:])
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			break
		}
		start += size
	}
	end := start
	for end < len(text) {
		r, size := utf8.DecodeRuneInString(text[end:])
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			break
		}
		end += size
	}

	return start, end
}

// managerAssertion verifies Service implements Manager.
var managerAssertion Manager = (*Service)(nil)
