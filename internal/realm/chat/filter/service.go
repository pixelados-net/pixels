// Package filter manages the immutable global chat word filter.
package filter

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"unicode/utf8"

	filterrepo "github.com/niflaot/pixels/internal/realm/chat/filter/repository"
	"github.com/niflaot/pixels/pkg/textfilter"
)

const (
	// MaxWordLength stores the maximum accepted filter word length.
	MaxWordLength = 32
)

var (
	// ErrInvalidWord reports malformed global filter input.
	ErrInvalidWord = errors.New("invalid global chat filter word")
)

// Snapshot stores an immutable global filter dictionary.
type Snapshot struct {
	// Words stores normalized words in stable order.
	Words []string
	// Matcher stores the immutable compiled automaton.
	Matcher *textfilter.Matcher
}

// Service manages persistent words and a lock-free read snapshot.
type Service struct {
	// store persists global filter words.
	store filterrepo.Store
	// snapshot stores the current immutable dictionary.
	snapshot atomic.Pointer[Snapshot]
}

// New creates a global chat filter service with an empty initial snapshot.
func New(store filterrepo.Store) *Service {
	service := &Service{store: store}
	service.snapshot.Store(&Snapshot{Matcher: textfilter.Compile(nil)})

	return service
}

// Refresh replaces the immutable dictionary from persistence.
func (service *Service) Refresh(ctx context.Context) error {
	words, err := service.store.List(ctx)
	if err != nil {
		return err
	}
	service.snapshot.Store(&Snapshot{Words: words, Matcher: textfilter.Compile(words)})

	return nil
}

// List returns a caller-owned global filter dictionary.
func (service *Service) List() []string {
	return append([]string(nil), service.snapshot.Load().Words...)
}

// Censor applies the immutable global dictionary without locks.
func (service *Service) Censor(text string) (string, bool) {
	return service.snapshot.Load().Matcher.Censor(text)
}

// Add persists one normalized word and refreshes the snapshot.
func (service *Service) Add(ctx context.Context, word string) error {
	word, err := normalize(word)
	if err != nil {
		return err
	}
	if err = service.store.Add(ctx, word); err != nil {
		return err
	}

	return service.Refresh(ctx)
}

// Remove deletes one normalized word and refreshes the snapshot.
func (service *Service) Remove(ctx context.Context, word string) error {
	word, err := normalize(word)
	if err != nil {
		return err
	}
	if err = service.store.Remove(ctx, word); err != nil {
		return err
	}

	return service.Refresh(ctx)
}

// normalize validates and normalizes a persisted whole word.
func normalize(word string) (string, error) {
	word = strings.ToLower(strings.TrimSpace(word))
	if word == "" || utf8.RuneCountInString(word) > MaxWordLength || strings.IndexFunc(word, func(r rune) bool { return r == ' ' || r == '\t' || r == '\n' || r == '\r' }) >= 0 {
		return "", ErrInvalidWord
	}

	return word, nil
}
